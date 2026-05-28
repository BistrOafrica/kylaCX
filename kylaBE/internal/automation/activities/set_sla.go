package activities

import (
	"context"
	"errors"
	"fmt"

	"kyla-be/shared/events"
)

// SetSLAActivity starts (or resets) the SLA timer on a conversation by
// delegating to the communication SLA engine. The engine selects the matching
// policy itself based on policy conditions; this activity is intentionally a
// thin trigger.
//
// Expected node.Config keys:
//
//	conversation_id string — required (or falls back to event.EntityID when the
//	                         trigger is a conversation event)
type SetSLAActivity struct {
	Deps Deps
}

func (a *SetSLAActivity) SetSLA(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if a.Deps.SLATimer == nil || a.Deps.ConversationStore == nil {
		return "", errors.New("set_sla: SLA timer or conversation store missing")
	}
	conversationID, _ := params["conversation_id"].(string)
	if conversationID == "" && event.Domain == "conversation" {
		conversationID = event.EntityID
	}
	if conversationID == "" {
		return "", errors.New("set_sla: conversation_id is required")
	}
	conv, err := a.Deps.ConversationStore.FindByID(conversationID, event.OrgID)
	if err != nil {
		return "", fmt.Errorf("set_sla: load conversation: %w", err)
	}
	if err := a.Deps.SLATimer.StartTimer(conv); err != nil {
		return "", fmt.Errorf("set_sla: start timer: %w", err)
	}
	return conv.ID, nil
}
