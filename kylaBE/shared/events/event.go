// Package events defines the shared domain event contract for all Kyla services.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NATS subject format: kyla.{org_id}.{domain}.{action}
const (
	SubjectObjectCreated = "kyla.%s.object.created"
	SubjectObjectUpdated = "kyla.%s.object.updated"
	SubjectObjectDeleted = "kyla.%s.object.deleted"

	SubjectTicketCreated  = "kyla.%s.ticket.created"
	SubjectTicketAssigned = "kyla.%s.ticket.assigned"
	SubjectTicketResolved = "kyla.%s.ticket.resolved"

	SubjectConversationCreated  = "kyla.%s.conversation.created"
	SubjectConversationResolved = "kyla.%s.conversation.resolved"
	SubjectMessageReceived      = "kyla.%s.conversation.message_received"

	SubjectCallStarted = "kyla.%s.call.started"
	SubjectCallEnded   = "kyla.%s.call.ended"

	SubjectDealCreated      = "kyla.%s.deal.created"
	SubjectDealStageChanged = "kyla.%s.deal.stage_changed"

	SubjectWorkflowTriggered = "kyla.%s.workflow.triggered"
	SubjectWorkflowCompleted = "kyla.%s.workflow.completed"
	SubjectWorkflowFailed    = "kyla.%s.workflow.failed"
)

// DomainEvent is the universal event envelope published to NATS.
type DomainEvent struct {
	ID            string          `json:"id"`
	OrgID         string          `json:"org_id"`
	WorkspaceID   string          `json:"workspace_id,omitempty"`
	Subject       string          `json:"subject"`
	Domain        string          `json:"domain"`
	Action        string          `json:"action"`
	EntityID      string          `json:"entity_id"`
	ActorID       string          `json:"actor_id,omitempty"`
	ActorType     string          `json:"actor_type,omitempty"`
	Payload       json.RawMessage `json:"payload"`
	OccurredAt    time.Time       `json:"occurred_at"`
	PublishedAt   time.Time       `json:"published_at"`
	SchemaVersion int             `json:"schema_version"`
}

// NewEvent constructs a DomainEvent with required fields.
func NewEvent(orgID, workspaceID, domain, action, entityID, actorID string, payload interface{}) (*DomainEvent, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &DomainEvent{
		ID:            uuid.New().String(),
		OrgID:         orgID,
		WorkspaceID:   workspaceID,
		Domain:        domain,
		Action:        action,
		EntityID:      entityID,
		ActorID:       actorID,
		Payload:       raw,
		OccurredAt:    now,
		PublishedAt:   now,
		SchemaVersion: 1,
	}, nil
}

// UnmarshalPayload decodes the event payload into target.
func (e *DomainEvent) UnmarshalPayload(target interface{}) error {
	return json.Unmarshal(e.Payload, target)
}

// Bus combines Publisher and Consumer.
type Bus interface {
	Publisher
	Consumer
}

// Publisher publishes domain events.
type Publisher interface {
	Publish(event *DomainEvent) error
}

// Consumer subscribes to domain events.
type Consumer interface {
	Subscribe(subject string, handler HandlerFunc) error
	QueueSubscribe(subject, queue string, handler HandlerFunc) error
}

// Subscription represents an active subscription that can be cancelled.
type Subscription interface {
	Unsubscribe() error
}

// StreamBus extends Consumer with raw subscriptions that return handles.
// Used for realtime streaming where clients need to unsubscribe on disconnect.
type StreamBus interface {
	SubscribeRaw(subject string, handler HandlerFunc) (Subscription, error)
}

// HandlerFunc is the callback for event consumers.
type HandlerFunc func(event *DomainEvent) error

// NoopBus is a no-op bus used during testing.
type NoopBus struct{}

func (n *NoopBus) Publish(_ *DomainEvent) error                    { return nil }
func (n *NoopBus) Subscribe(_ string, _ HandlerFunc) error         { return nil }
func (n *NoopBus) QueueSubscribe(_, _ string, _ HandlerFunc) error { return nil }

// noopSubscription is a no-op subscription.
type noopSubscription struct{}

func (n *noopSubscription) Unsubscribe() error { return nil }

// SubscribeRaw no-op implementation.
func (n *NoopBus) SubscribeRaw(_ string, _ HandlerFunc) (Subscription, error) {
	return &noopSubscription{}, nil
}
