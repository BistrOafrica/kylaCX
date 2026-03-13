package events

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	nats "github.com/nats-io/nats.go"
)

// NatsBus is the production event bus backed by NATS JetStream.
type NatsBus struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewNatsBus connects to NATS and returns a configured bus.
func NewNatsBus(url string) (*NatsBus, error) {
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("[NATS] disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[NATS] reconnected to %s", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}
	bus := &NatsBus{conn: nc, js: js}
	if err := bus.ensureStream(); err != nil {
		return nil, err
	}
	log.Printf("[NATS] connected to %s", nc.ConnectedUrl())
	return bus, nil
}

func (b *NatsBus) ensureStream() error {
	const name = "KYLA"
	_, err := b.js.StreamInfo(name)
	if err == nats.ErrStreamNotFound {
		_, err = b.js.AddStream(&nats.StreamConfig{
			Name:      name,
			Subjects:  []string{"kyla.>"},
			MaxAge:    7 * 24 * time.Hour,
			Storage:   nats.FileStorage,
			Replicas:  1,
			Retention: nats.LimitsPolicy,
		})
		if err != nil {
			return fmt.Errorf("create KYLA stream: %w", err)
		}
		log.Println("[NATS] JetStream stream 'KYLA' created")
	} else if err != nil {
		return fmt.Errorf("stream info: %w", err)
	}
	return nil
}

// Publish encodes a DomainEvent and publishes it to NATS JetStream.
func (b *NatsBus) Publish(event *DomainEvent) error {
	subject := fmt.Sprintf("kyla.%s.%s.%s", event.OrgID, event.Domain, event.Action)
	event.Subject = subject
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if _, err := b.js.Publish(subject, data); err != nil {
		return fmt.Errorf("nats publish [%s]: %w", subject, err)
	}
	return nil
}

// Subscribe registers a push subscription on a NATS subject pattern.
func (b *NatsBus) Subscribe(subject string, handler HandlerFunc) error {
	_, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		ev := &DomainEvent{}
		if err := json.Unmarshal(msg.Data, ev); err != nil {
			log.Printf("[NATS] unmarshal error: %v", err)
			_ = msg.Nak()
			return
		}
		if err := handler(ev); err != nil {
			log.Printf("[NATS] handler error: %v", err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	return err
}

// QueueSubscribe registers a competing consumer group.
func (b *NatsBus) QueueSubscribe(subject, queue string, handler HandlerFunc) error {
	_, err := b.js.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		ev := &DomainEvent{}
		if err := json.Unmarshal(msg.Data, ev); err != nil {
			log.Printf("[NATS] unmarshal error [q=%s]: %v", queue, err)
			_ = msg.Nak()
			return
		}
		if err := handler(ev); err != nil {
			log.Printf("[NATS] handler error [q=%s]: %v", queue, err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	})
	return err
}

// natsSubscription wraps a NATS JetStream subscription to satisfy Subscription interface.
type natsSubscription struct {
	sub *nats.Subscription
}

func (s *natsSubscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
}

// SubscribeRaw creates a raw NATS subscription without JetStream durable consumer.
// Returns a Subscription handle that the caller MUST call Unsubscribe() on to clean up.
func (b *NatsBus) SubscribeRaw(subject string, handler HandlerFunc) (Subscription, error) {
	sub, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		var event DomainEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("[nats] unmarshal error on subject=%s: %v", subject, err)
			_ = msg.Nak()
			return
		}
		if err := handler(&event); err != nil {
			log.Printf("[nats] handler error on subject=%s event=%s: %v", subject, event.ID, err)
			_ = msg.Nak()
		} else {
			_ = msg.Ack()
		}
	})
	if err != nil {
		return nil, err
	}
	return &natsSubscription{sub: sub}, nil
}

// Close drains and closes the NATS connection.
func (b *NatsBus) Close() {
	if b.conn != nil {
		_ = b.conn.Drain()
	}
}
