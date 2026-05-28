package activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"kyla-be/internal/communication"
	"kyla-be/shared/events"
)

// SendMessageActivity dispatches an outbound message on an existing
// conversation. The channel adapter (WA/email/SMS/voice/webchat) is resolved
// from conversation.Channel.
//
// Expected node.Config keys:
//
//	conversation_id string — required (or falls back to event.EntityID when
//	                         the trigger is a conversation event)
//	content_type    string — optional, defaults to "text"
//	content         any    — required; serialised into Message.Content JSONB
//	sender_id       string — optional, defaults to "" (bot/system message)
//	sender_type     string — optional, defaults to "bot"
type SendMessageActivity struct {
	Deps Deps
}

func (a *SendMessageActivity) SendMessage(ctx context.Context, params map[string]interface{}, event events.DomainEvent) (string, error) {
	if a.Deps.ConversationStore == nil || a.Deps.MessageStore == nil || a.Deps.AdapterRegistry == nil {
		return "", errors.New("send_message: communication dependencies missing")
	}
	conversationID, _ := params["conversation_id"].(string)
	if conversationID == "" && event.Domain == "conversation" {
		conversationID = event.EntityID
	}
	if conversationID == "" {
		return "", errors.New("send_message: conversation_id is required")
	}
	contentRaw, ok := params["content"]
	if !ok {
		return "", errors.New("send_message: content is required")
	}
	contentBytes, err := json.Marshal(contentRaw)
	if err != nil {
		return "", fmt.Errorf("send_message: marshal content: %w", err)
	}

	conv, err := a.Deps.ConversationStore.FindByID(conversationID, event.OrgID)
	if err != nil {
		return "", fmt.Errorf("send_message: load conversation: %w", err)
	}

	contentType, _ := params["content_type"].(string)
	if contentType == "" {
		contentType = "text"
	}
	senderType, _ := params["sender_type"].(string)
	if senderType == "" {
		senderType = "bot"
	}
	senderID, _ := params["sender_id"].(string)

	msg := &communication.Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		SenderType:     senderType,
		Channel:        conv.Channel,
		ContentType:    contentType,
		Content:        contentBytes,
		Status:         "pending",
	}
	created, err := a.Deps.MessageStore.Create(msg)
	if err != nil {
		return "", fmt.Errorf("send_message: persist message: %w", err)
	}

	if err := a.Deps.AdapterRegistry.Dispatch(ctx, conv, created); err != nil {
		_ = a.Deps.MessageStore.UpdateStatus(created.ID, "failed")
		return "", fmt.Errorf("send_message: dispatch via %s: %w", conv.Channel, err)
	}
	_ = a.Deps.MessageStore.UpdateStatus(created.ID, "sent")
	return created.ID, nil
}
