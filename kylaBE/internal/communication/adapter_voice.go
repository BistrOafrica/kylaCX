package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"kyla-be/internal/objectcore"
	"kyla-be/shared/events"
)

// VoiceCallBridge listens for call.ended NATS events and creates conversation records.
type VoiceCallBridge struct {
	convStore *ConversationStore
	msgStore  *MessageStore
	ocStore   ObjectCoreGateway
	eventBus  events.Publisher
	consumer  events.Consumer
}

// NewVoiceCallBridge constructs a VoiceCallBridge.
func NewVoiceCallBridge(
	convStore *ConversationStore,
	msgStore *MessageStore,
	ocStore ObjectCoreGateway,
	eventBus events.Publisher,
	consumer events.Consumer,
) *VoiceCallBridge {
	return &VoiceCallBridge{
		convStore: convStore,
		msgStore:  msgStore,
		ocStore:   ocStore,
		eventBus:  eventBus,
		consumer:  consumer,
	}
}

// Start subscribes to call.ended events and processes them.
func (b *VoiceCallBridge) Start(ctx context.Context) error {
	subject := "kyla.*.call.ended"
	if err := b.consumer.QueueSubscribe(subject, "voice-bridge", b.handleCallEnded); err != nil {
		return fmt.Errorf("subscribe to %s: %w", subject, err)
	}
	log.Printf("[voice] subscribed to %s (queue: voice-bridge)", subject)
	<-ctx.Done()
	log.Println("[voice] bridge stopped")
	return nil
}

func (b *VoiceCallBridge) handleCallEnded(event *events.DomainEvent) error {
	var payload struct {
		CallSID         string `json:"call_sid"`
		From            string `json:"from"`
		To              string `json:"to"`
		DurationSeconds int    `json:"duration_seconds"`
		Direction       string `json:"direction"`
		OrgID           string `json:"org_id"`
		WorkspaceID     string `json:"workspace_id"`
		ContactID       string `json:"contact_id"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("[voice] unmarshal call.ended payload error: %v", err)
		return err
	}

	if payload.CallSID == "" || payload.OrgID == "" || payload.WorkspaceID == "" {
		log.Println("[voice] call.ended missing required fields")
		return nil
	}

	// Find or create conversation using call_sid as channel_ref
	conv, err := b.convStore.FindByChannelRef(payload.OrgID, ChannelVoice, payload.CallSID)
	if err != nil {
		log.Printf("[voice] lookup conversation error: %v", err)
		return err
	}
	if conv == nil {
		// Create new conversation
		conv = &Conversation{
			OrgID:       payload.OrgID,
			WorkspaceID: payload.WorkspaceID,
			Channel:     ChannelVoice,
			ChannelRef:  payload.CallSID,
			ContactID:   payload.ContactID,
			Status:      StatusResolved, // calls are resolved immediately
			Priority:    PriorityNormal,
			Subject:     fmt.Sprintf("Call from %s", payload.From),
			Meta:        []byte(`{}`),
		}

		conv, err = b.convStore.Create(conv)
		if err != nil {
			log.Printf("[voice] create conversation error: %v", err)
			return err
		}

		// Create Object Core record
		ocObj := &objectcore.Object{
			OrgID:       conv.OrgID,
			WorkspaceID: conv.WorkspaceID,
			TypeSlug:    "conversation",
			Data:        []byte(fmt.Sprintf(`{"channel":"voice","call_sid":%q}`, payload.CallSID)),
			CreatedBy:   "system",
		}
		ocObj.ID = conv.ID
		_, _ = b.ocStore.CreateObject(ocObj, "system")

		b.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "created", "", conv)
	}

	// Create call summary message
	content, _ := json.Marshal(map[string]interface{}{
		"type":             "call_summary",
		"duration_seconds": payload.DurationSeconds,
		"direction":        payload.Direction,
		"from":             payload.From,
		"to":               payload.To,
	})

	msg := &Message{
		ConversationID: conv.ID,
		SenderType:     SenderSystem,
		Channel:        ChannelVoice,
		ContentType:    ContentTypeText,
		Content:        content,
		Status:         MsgStatusReceived,
		ExternalID:     payload.CallSID,
	}

	created, err := b.msgStore.Create(msg)
	if err != nil {
		log.Printf("[voice] create message error: %v", err)
		return err
	}

	// Append Object Core timeline event
	_ = b.ocStore.AppendEvent(&objectcore.ObjectEvent{
		OrgID:     conv.OrgID,
		ObjectID:  conv.ID,
		ActorID:   "system",
		ActorType: "system",
		EventType: "call_ended",
		Payload:   content,
	})

	b.publishEvent(conv.OrgID, conv.WorkspaceID, conv.ID, "message_received", "", created)
	log.Printf("[voice] processed call.ended call_sid=%s conv=%s msg=%s", payload.CallSID, conv.ID, created.ID)

	return nil
}

func (b *VoiceCallBridge) publishEvent(orgID, workspaceID, entityID, action, actorID string, payload interface{}) {
	ev, err := events.NewEvent(orgID, workspaceID, "conversation", action, entityID, actorID, payload)
	if err != nil {
		log.Printf("[voice] event build error (action=%s entity=%s): %v", action, entityID, err)
		return
	}
	if err := b.eventBus.Publish(ev); err != nil {
		log.Printf("[voice] event publish error (action=%s entity=%s): %v", action, entityID, err)
	}
}
