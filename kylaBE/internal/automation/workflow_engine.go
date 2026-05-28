package automation

import (
	"fmt"
	"time"

	"kyla-be/internal/automation/activities"
	"kyla-be/shared/events"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ActionType enumerates the built-in workflow action node types.
// Each value maps to either an inline workflow primitive (delay,
// start_workflow) or a registered Temporal Activity (everything else).
const (
	ActionTypeDelay            = "delay"
	ActionTypeStartWorkflow    = "start_workflow"
	ActionTypeUpdateObject     = "update_object"
	ActionTypeAssignUser       = "assign_user"
	ActionTypeCreateObject     = "create_object"
	ActionTypeCreateTask       = "create_task"
	ActionTypeSendMessage      = "send_message"
	ActionTypeInvokeWebhook    = "invoke_webhook"
	ActionTypeSetSLA           = "set_sla"
	ActionTypeSendNotification = "send_notification"
	ActionTypeRunAISkill       = "run_ai_skill"
)

// WorkflowInput is the payload AutomationWorkflow accepts. The trigger event
// is captured up-front so all activities see the same immutable snapshot, and
// the action list is passed in (rather than re-fetched) so workflow logic stays
// deterministic across replays.
type WorkflowInput struct {
	WorkflowID    string             `json:"workflow_id"`
	TemporalRunID string             `json:"temporal_run_id"`
	Actions       []ActionNode       `json:"actions"`
	Trigger       events.DomainEvent `json:"trigger"`
}

// AutomationWorkflow is the top-level Temporal workflow function.
// It walks the action list and dispatches each node to either an inline
// workflow primitive or a registered activity.
//
// Versioning: behaviour changes that affect determinism must be gated with
// workflow.GetVersion() per the architectural decision recorded in the
// roadmap. In-flight runs replay against their original version.
func AutomationWorkflow(ctx workflow.Context, in WorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("automation workflow started",
		"workflow_id", in.WorkflowID,
		"trigger_subject", in.Trigger.Subject,
		"actions", len(in.Actions),
	)

	for i, node := range in.Actions {
		logger.Info("executing action", "index", i, "type", node.Type, "id", node.ID)

		// Per-node configured delay (separate from the "delay" action type).
		// Useful for staggering actions inside a single workflow without
		// requiring a dedicated delay node between every step.
		if node.DelaySeconds > 0 {
			_ = workflow.Sleep(ctx, time.Duration(node.DelaySeconds)*time.Second)
		}

		if err := executeAction(ctx, node, in.Trigger); err != nil {
			logger.Error("action failed", "index", i, "type", node.Type, "error", err)
			// Per-node policy: an explicit on_failure jump short-circuits the
			// rest of the workflow. Without one, the workflow halts on first
			// error — matches the principle of least surprise for now.
			// (Per-workflow "continue on failure" can be added later.)
			return fmt.Errorf("action %d (%s): %w", i, node.Type, err)
		}
	}

	logger.Info("automation workflow completed", "workflow_id", in.WorkflowID)
	return nil
}

// executeAction dispatches a single ActionNode. Delay and start_workflow are
// handled inline (workflow.Sleep and workflow.ExecuteChildWorkflow are durable
// primitives that don't need an activity round-trip).
func executeAction(ctx workflow.Context, node ActionNode, trigger events.DomainEvent) error {
	switch node.Type {
	case ActionTypeDelay:
		dur := durationFromConfig(node.Config)
		if dur <= 0 {
			return fmt.Errorf("delay: invalid or missing duration_seconds")
		}
		return workflow.Sleep(ctx, dur)

	case ActionTypeStartWorkflow:
		return executeStartWorkflow(ctx, node, trigger)

	case ActionTypeUpdateObject:
		return runActivity(ctx, (*activities.UpdateObjectActivity).UpdateObject, node, trigger)
	case ActionTypeAssignUser:
		return runActivity(ctx, (*activities.AssignUserActivity).AssignUser, node, trigger)
	case ActionTypeCreateObject:
		return runActivity(ctx, (*activities.CreateObjectActivity).CreateObject, node, trigger)
	case ActionTypeCreateTask:
		return runActivity(ctx, (*activities.CreateTaskActivity).CreateTask, node, trigger)
	case ActionTypeSendMessage:
		return runActivity(ctx, (*activities.SendMessageActivity).SendMessage, node, trigger)
	case ActionTypeInvokeWebhook:
		return runActivity(ctx, (*activities.InvokeWebhookActivity).InvokeWebhook, node, trigger)
	case ActionTypeSetSLA:
		return runActivity(ctx, (*activities.SetSLAActivity).SetSLA, node, trigger)
	case ActionTypeSendNotification:
		return runActivity(ctx, (*activities.SendNotificationActivity).SendNotification, node, trigger)
	case ActionTypeRunAISkill:
		return runActivity(ctx, (*activities.RunAISkillActivity).RunAISkill, node, trigger)

	default:
		// Unknown types are non-fatal: the engine logs and continues so a
		// stale UI submitting an unknown action type doesn't tank the run.
		workflow.GetLogger(ctx).Warn("unknown action type, skipping", "type", node.Type)
		return nil
	}
}

// runActivity is the common shape for dispatching an activity registered as a
// method on a per-activity struct. The activity function signature is always
// (ctx, map[string]interface{}, events.DomainEvent) -> (string, error).
func runActivity(ctx workflow.Context, fn interface{}, node ActionNode, trigger events.DomainEvent) error {
	opts := activityOptionsForNode(node)
	var result string
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, opts),
		fn,
		node.Config,
		trigger,
	).Get(ctx, &result)
}

// executeStartWorkflow chains into another workflow as a child execution.
// The child runs in the same task queue so dispatch is local.
//
// Expected node.Config keys:
//
//	workflow_id  string — required; the child workflow definition ID
//	wait         bool   — optional; when true, block until child completes
type startWorkflowChildInput struct {
	WorkflowID string             `json:"workflow_id"`
	Actions    []ActionNode       `json:"actions"`
	Trigger    events.DomainEvent `json:"trigger"`
}

func executeStartWorkflow(ctx workflow.Context, node ActionNode, trigger events.DomainEvent) error {
	childID, _ := node.Config["workflow_id"].(string)
	if childID == "" {
		return fmt.Errorf("start_workflow: workflow_id is required")
	}
	// The child workflow body is the same AutomationWorkflow function; we just
	// reuse it. The child's action list must be fetched at dispatch time —
	// but doing so inside a workflow would be non-deterministic. So instead
	// we expect callers to inline the child's actions in node.Config["actions"]
	// (the visual builder will hydrate this at workflow-save time).
	actionsRaw, _ := node.Config["actions"].([]interface{})
	childActions := make([]ActionNode, 0, len(actionsRaw))
	for _, a := range actionsRaw {
		if m, ok := a.(map[string]interface{}); ok {
			node := ActionNode{}
			if t, ok := m["type"].(string); ok {
				node.Type = t
			}
			if cfg, ok := m["config"].(map[string]interface{}); ok {
				node.Config = cfg
			}
			childActions = append(childActions, node)
		}
	}

	childInput := WorkflowInput{
		WorkflowID: childID,
		Actions:    childActions,
		Trigger:    trigger,
	}

	wait, _ := node.Config["wait"].(bool)
	future := workflow.ExecuteChildWorkflow(ctx, AutomationWorkflow, childInput)
	if wait {
		return future.Get(ctx, nil)
	}
	return nil
}

// activityOptionsForNode allows per-node timeout / retry overrides via
// node.Config["timeout_seconds"] and node.Config["max_attempts"]. Falls back
// to defaultActivityOptions when no overrides are present.
func activityOptionsForNode(node ActionNode) workflow.ActivityOptions {
	opts := defaultActivityOptions()
	if v, ok := node.Config["timeout_seconds"]; ok {
		switch x := v.(type) {
		case float64:
			opts.StartToCloseTimeout = time.Duration(x) * time.Second
		case int:
			opts.StartToCloseTimeout = time.Duration(x) * time.Second
		}
	}
	if v, ok := node.Config["max_attempts"]; ok {
		if rp := opts.RetryPolicy; rp != nil {
			switch x := v.(type) {
			case float64:
				rp.MaximumAttempts = int32(x)
			case int:
				rp.MaximumAttempts = int32(x)
			}
		}
	}
	return opts
}

// defaultActivityOptions provides a sane retry/timeout policy for activities
// that don't override per-node. Overrides come from node.Config in step 6+.
func defaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
}

// durationFromConfig accepts either an integer seconds value or a string
// like "5s" / "1h30m" — both shapes are common in JSON-authored workflows.
func durationFromConfig(cfg map[string]interface{}) time.Duration {
	if cfg == nil {
		return 0
	}
	if v, ok := cfg["duration_seconds"]; ok {
		switch x := v.(type) {
		case float64:
			return time.Duration(x) * time.Second
		case int:
			return time.Duration(x) * time.Second
		case int64:
			return time.Duration(x) * time.Second
		}
	}
	if v, ok := cfg["duration"]; ok {
		if s, ok := v.(string); ok {
			if d, err := time.ParseDuration(s); err == nil {
				return d
			}
		}
	}
	return 0
}
