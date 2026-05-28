package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kyla-be/shared/events"

	"go.temporal.io/sdk/client"
)

// Executor wraps the Temporal client + store with the start-workflow logic
// shared by the NATS consumer and the TestRunWorkflow gRPC handler.
//
// The Executor is responsible for:
//  1. Creating a workflow_runs projection row before kicking off Temporal.
//  2. Building the deterministic WorkflowID (so duplicate events from NATS
//     redelivery collapse into a single Temporal run).
//  3. Calling client.ExecuteWorkflow on the configured task queue.
type Executor struct {
	temporal  client.Client
	store     *Store
	taskQueue string
}

func NewExecutor(temporal client.Client, store *Store, taskQueue string) *Executor {
	return &Executor{temporal: temporal, store: store, taskQueue: taskQueue}
}

// Enabled reports whether Temporal is available. Callers (consumer, gRPC test
// run) use this to short-circuit early with a meaningful error.
func (e *Executor) Enabled() bool { return e != nil && e.temporal != nil }

// StartWorkflow kicks off a single AutomationWorkflow execution for the given
// workflow definition and trigger event. Returns the Temporal RunID so callers
// can return it to the client (gRPC TestRun) or log it (consumer).
func (e *Executor) StartWorkflow(ctx context.Context, wf *Workflow, event *events.DomainEvent) (string, error) {
	if !e.Enabled() {
		return "", fmt.Errorf("temporal client not configured; automation disabled")
	}
	if wf == nil || event == nil {
		return "", fmt.Errorf("workflow and event are required")
	}

	// Deterministic WorkflowID: dedupes NATS redeliveries naturally because
	// Temporal rejects duplicate IDs by default.
	workflowID := fmt.Sprintf("wf-%s-%s", wf.ID, event.ID)

	input := WorkflowInput{
		WorkflowID: wf.ID,
		Actions:    wf.Actions,
		Trigger:    *event,
	}

	run, err := e.temporal.ExecuteWorkflow(ctx,
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: e.taskQueue,
		},
		AutomationWorkflow, input,
	)
	if err != nil {
		return "", fmt.Errorf("execute workflow: %w", err)
	}

	// Insert the projection row up-front so the UI sees a "running" run
	// immediately. Failures here are non-fatal — the workflow itself proceeds.
	contextBytes, _ := json.Marshal(map[string]interface{}{
		"event_subject": event.Subject,
		"event_domain":  event.Domain,
		"event_action":  event.Action,
		"entity_id":     event.EntityID,
	})
	startedAt := time.Now().UTC()
	if _, err := e.store.CreateRun(&WorkflowRun{
		WorkflowID:     wf.ID,
		TemporalRunID:  run.GetRunID(),
		OrgID:          wf.OrgID,
		TriggerEventID: event.ID,
		Status:         RunStatusRunning,
		Context:        contextBytes,
		StartedAt:      &startedAt,
	}); err != nil {
		// Log via the standard logger to avoid coupling to a domain-specific
		// logger — the projection is best-effort.
		fmt.Printf("[automation] warn: create run projection failed: %v\n", err)
	}

	// Bump the run counter on the workflow definition.
	_ = e.store.IncrementRunCount(wf.ID)

	return run.GetRunID(), nil
}
