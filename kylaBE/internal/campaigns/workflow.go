package campaigns

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// CampaignExecutionInput is the payload CampaignExecutionWorkflow accepts.
// Captured up-front so all activities run against the same snapshot — the
// campaign row may change during execution but the workflow's view does not.
type CampaignExecutionInput struct {
	CampaignID  string                 `json:"campaign_id"`
	OrgID       string                 `json:"org_id"`
	WorkspaceID string                 `json:"workspace_id"`
	Channel     string                 `json:"channel"`
	Payload     map[string]interface{} `json:"payload"`
	// StartAt is honored only for scheduled_once. The workflow Sleeps until
	// then before resolving the audience and fanning out.
	StartAt time.Time `json:"start_at,omitempty"`
}

// CampaignExecutionWorkflow drives a single campaign run end-to-end:
//
//  1. Sleep until StartAt (for scheduled_once).
//  2. Resolve the audience into campaign_recipients rows.
//  3. Fan out one SendRecipient activity per recipient, with per-channel
//     concurrency capped to keep provider rate-limits happy.
//  4. Finalise — recompute stats, mark campaign complete.
//
// Recurring schedules are NOT handled here — they're driven by Temporal
// Schedules which trigger one CampaignExecutionWorkflow per tick.
func CampaignExecutionWorkflow(ctx workflow.Context, in CampaignExecutionInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("campaign workflow started",
		"campaign_id", in.CampaignID,
		"channel", in.Channel,
	)

	// 1. Honor scheduled_once start time. Sleeps that would be in the past
	//    return immediately, so unconditional Sleep is safe.
	if !in.StartAt.IsZero() {
		delay := workflow.Now(ctx).Sub(in.StartAt)
		if delay < 0 {
			_ = workflow.Sleep(ctx, -delay)
		}
	}

	// 2. Resolve audience. Activity is idempotent so a retried workflow does
	//    not duplicate recipient rows.
	resolveOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	var resolvedCount int
	if err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, resolveOpts),
		(*ResolveAudienceActivity).ResolveAudience,
		in.CampaignID,
	).Get(ctx, &resolvedCount); err != nil {
		return fmt.Errorf("resolve audience: %w", err)
	}
	logger.Info("audience resolved", "count", resolvedCount)

	if resolvedCount == 0 {
		// Nothing to send. Mark complete and exit.
		return finalise(ctx, in.CampaignID, string(StatusCompleted))
	}

	// 3. Fan out sends. Pull recipients in pages (the store caps at 1000 per
	//    call) and dispatch each as an activity. Concurrency is bounded by
	//    the worker's MaxConcurrentActivityExecutionSize knob; per-workflow
	//    we fire-and-collect to keep history compact.
	sendOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    4,
		},
	}

	// Page through queued recipients until none remain.
	const pageSize = 500
	for {
		var batch []SendRecipientInput
		if err := workflow.ExecuteActivity(
			workflow.WithActivityOptions(ctx, resolveOpts),
			loadSendBatchActivityName,
			loadSendBatchInput{CampaignID: in.CampaignID, Limit: pageSize, Channel: in.Channel, WorkspaceID: in.WorkspaceID, Payload: in.Payload},
		).Get(ctx, &batch); err != nil {
			return fmt.Errorf("load send batch: %w", err)
		}
		if len(batch) == 0 {
			break
		}

		futures := make([]workflow.Future, 0, len(batch))
		for _, item := range batch {
			f := workflow.ExecuteActivity(
				workflow.WithActivityOptions(ctx, sendOpts),
				(*SendRecipientActivity).SendRecipient,
				item,
			)
			futures = append(futures, f)
		}
		// Drain — we want to capture failures in logs but proceed regardless.
		// The activity itself already records per-recipient failure into the DB.
		for _, f := range futures {
			if err := f.Get(ctx, nil); err != nil {
				logger.Warn("send recipient failed (recorded on row)", "error", err)
			}
		}
	}

	// 4. Finalise.
	return finalise(ctx, in.CampaignID, string(StatusCompleted))
}

// loadSendBatchActivityName is the registered name of the LoadSendBatchActivity
// method below. Declared as a string constant so the workflow body doesn't
// reference the receiver type directly — keeps workflow imports minimal.
const loadSendBatchActivityName = "LoadSendBatch"

// LoadSendBatchActivity hydrates the next page of queued recipients into the
// fan-out input shape. Pulling DB rows from within the workflow body would be
// non-deterministic, so this lives behind an activity.
type LoadSendBatchActivity struct {
	Deps ActivityDeps
}

type loadSendBatchInput struct {
	CampaignID  string                 `json:"campaign_id"`
	Limit       int                    `json:"limit"`
	Channel     string                 `json:"channel"`
	WorkspaceID string                 `json:"workspace_id"`
	Payload     map[string]interface{} `json:"payload"`
}

func (a *LoadSendBatchActivity) LoadSendBatch(_ context.Context, in loadSendBatchInput) ([]SendRecipientInput, error) {
	if a.Deps.Store == nil {
		return nil, fmt.Errorf("LoadSendBatch: store missing")
	}
	rows, err := a.Deps.Store.ListQueuedRecipients(in.CampaignID, in.Limit)
	if err != nil {
		return nil, fmt.Errorf("LoadSendBatch: %w", err)
	}
	out := make([]SendRecipientInput, 0, len(rows))
	for _, r := range rows {
		out = append(out, SendRecipientInput{
			RecipientID: r.ID,
			CampaignID:  r.CampaignID,
			OrgID:       r.OrgID,
			WorkspaceID: in.WorkspaceID,
			Channel:     in.Channel,
			ContactRef:  r.ContactRef,
			Payload:     in.Payload,
		})
	}
	return out, nil
}

func finalise(ctx workflow.Context, campaignID, terminal string) error {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	return workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, opts),
		(*FinaliseCampaignActivity).FinaliseCampaign,
		campaignID, terminal,
	).Get(ctx, nil)
}
