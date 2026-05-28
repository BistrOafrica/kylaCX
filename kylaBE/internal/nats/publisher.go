package nats

import (
	"encoding/json"
	"fmt"
)

// Standard event subjects.
const (
	SubjectOrgCreated    = "org.created"
	SubjectUserCreated   = "user.created"
	SubjectAuthLogin     = "auth.login"
	SubjectInviteSent    = "invitation.sent"
	SubjectInviteCreated = "invitation.created"
)

// EventPublisher publishes domain events to NATS subjects.
type EventPublisher struct {
	client *Client
}

// NewEventPublisher creates an EventPublisher backed by the given Client.
func NewEventPublisher(c *Client) *EventPublisher {
	return &EventPublisher{client: c}
}

// Publish sends raw bytes to the given subject.
func (p *EventPublisher) Publish(subject string, payload []byte) error {
	if err := p.client.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("nats publish %s: %w", subject, err)
	}
	return nil
}

// PublishJSON marshals v to JSON and publishes it to subject.
func (p *EventPublisher) PublishJSON(subject string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("nats publishJSON marshal: %w", err)
	}
	return p.Publish(subject, data)
}
