package campaigns

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
)

// Executor implements server.Launcher by talking to Temporal. The server
// depends only on the interface so a nil Executor means "Temporal not
// available" and CRUD continues to work.
type Executor struct {
	temporal  client.Client
	taskQueue string
}

func NewExecutor(temporal client.Client, taskQueue string) *Executor {
	return &Executor{temporal: temporal, taskQueue: taskQueue}
}

// Enabled mirrors automation.Executor — used by the server to skip lifecycle
// RPCs cleanly when Temporal isn't dialed.
func (e *Executor) Enabled() bool { return e != nil && e.temporal != nil }

// Launch starts the workflow for one-shot campaigns and (for recurring) creates
// a Temporal Schedule that fires the same workflow on cron.
// Returns the Temporal workflow ID (always set for one-shot) and the schedule
// ID (set only for recurring).
func (e *Executor) Launch(ctx context.Context, c *Campaign) (string, string, error) {
	if !e.Enabled() {
		return "", "", fmt.Errorf("temporal not available")
	}
	sched, err := DecodeSchedule(c.Schedule)
	if err != nil {
		return "", "", fmt.Errorf("decode schedule: %w", err)
	}

	in := CampaignExecutionInput{
		CampaignID:  c.ID,
		OrgID:       c.OrgID,
		WorkspaceID: c.WorkspaceID,
		Channel:     c.Channel,
		Payload:     decodePayload(c.Payload),
		StartAt:     ScheduleStartTime(sched),
	}

	switch sched.Mode {
	case ScheduleRecurring:
		scheduleID := fmt.Sprintf("campaign-schedule-%s", c.ID)
		_, err := e.temporal.ScheduleClient().Create(ctx, client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{sched.Cron},
				TimeZoneName:    sched.Timezone,
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        fmt.Sprintf("campaign-%s-{{.ScheduledTime.Unix}}", c.ID),
				Workflow:  CampaignExecutionWorkflow,
				Args:      []interface{}{in},
				TaskQueue: e.taskQueue,
			},
		})
		if err != nil {
			return "", "", fmt.Errorf("create schedule: %w", err)
		}
		log.Printf("[campaigns] created schedule %s for campaign %s (cron=%s tz=%s)",
			scheduleID, c.ID, sched.Cron, sched.Timezone)
		return "", scheduleID, nil

	default:
		// immediate + scheduled_once both run a single workflow.
		workflowID := fmt.Sprintf("campaign-%s", c.ID)
		run, err := e.temporal.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: e.taskQueue,
		}, CampaignExecutionWorkflow, in)
		if err != nil {
			return "", "", fmt.Errorf("execute workflow: %w", err)
		}
		log.Printf("[campaigns] started workflow %s run=%s for campaign %s", workflowID, run.GetRunID(), c.ID)
		return workflowID, "", nil
	}
}

// Pause halts further sends. For one-shot it cancels the workflow; for
// recurring it pauses the schedule (so already-running ticks complete).
func (e *Executor) Pause(ctx context.Context, c *Campaign) error {
	if !e.Enabled() {
		return fmt.Errorf("temporal not available")
	}
	if c.TemporalScheduleID != "" {
		h := e.temporal.ScheduleClient().GetHandle(ctx, c.TemporalScheduleID)
		return h.Pause(ctx, client.SchedulePauseOptions{Note: "paused via CampaignService.PauseCampaign"})
	}
	if c.TemporalWorkflowID != "" {
		return e.temporal.CancelWorkflow(ctx, c.TemporalWorkflowID, "")
	}
	return nil
}

// Cancel terminates the run and (for recurring) deletes the schedule.
func (e *Executor) Cancel(ctx context.Context, c *Campaign) error {
	if !e.Enabled() {
		return fmt.Errorf("temporal not available")
	}
	if c.TemporalScheduleID != "" {
		h := e.temporal.ScheduleClient().GetHandle(ctx, c.TemporalScheduleID)
		if err := h.Delete(ctx); err != nil {
			return fmt.Errorf("delete schedule: %w", err)
		}
	}
	if c.TemporalWorkflowID != "" {
		// Terminate (not cancel) so in-flight activities get a hard stop.
		return e.temporal.TerminateWorkflow(ctx, c.TemporalWorkflowID, "", "cancelled via CampaignService.CancelCampaign")
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func decodePayload(raw []byte) map[string]interface{} {
	out := map[string]interface{}{}
	if len(raw) == 0 {
		return out
	}
	// Workflow inputs must be JSON-serialisable; the raw JSONB already is.
	// We unmarshal into a map so activities see plain Go types.
	_ = json.Unmarshal(raw, &out)
	return out
}
