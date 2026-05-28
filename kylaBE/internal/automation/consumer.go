package automation

import (
	"context"
	"fmt"
	"log"

	"kyla-be/shared/events"
)

// ConsumerQueueGroup is the NATS queue group name. All server instances using
// this group share inbound events — exactly one consumer handles each message,
// which prevents duplicate workflow starts across a horizontally-scaled fleet.
const ConsumerQueueGroup = "kyla-automation"

// ConsumerSubject is the wildcard subject the consumer listens on.
// "kyla.*.>" matches every domain event regardless of org or sub-topic.
const ConsumerSubject = "kyla.*.>"

// Consumer subscribes to the platform event stream and kicks off matching
// workflows via the Executor. It does NOT evaluate fine-grained conditions
// (workflow.Conditions) — that work happens inside the workflow body, after
// the run is persisted, so failed condition matches are still visible in the
// run history.
type Consumer struct {
	store    *Store
	executor *Executor
	bus      events.Bus
	sub      events.Subscription
}

func NewConsumer(store *Store, executor *Executor, bus events.Bus) *Consumer {
	return &Consumer{store: store, executor: executor, bus: bus}
}

// Start subscribes to the platform event stream. Safe to call when Temporal
// is unavailable — the consumer still subscribes but skips dispatch.
//
// Returns an error if the underlying bus subscription fails. NATS unavailable
// is not an error here; the NoopBus simply never delivers events.
func (c *Consumer) Start(ctx context.Context) error {
	if c.bus == nil {
		return fmt.Errorf("automation consumer: event bus is nil")
	}
	if err := c.bus.QueueSubscribe(ConsumerSubject, ConsumerQueueGroup, c.handle); err != nil {
		return fmt.Errorf("automation consumer subscribe: %w", err)
	}
	log.Printf("automation consumer: subscribed to %q (queue=%q)", ConsumerSubject, ConsumerQueueGroup)
	return nil
}

// Stop tears down the subscription. Best-effort.
func (c *Consumer) Stop() {
	if c.sub != nil {
		_ = c.sub.Unsubscribe()
		c.sub = nil
	}
}

// handle is the NATS message callback. It matches the event against persisted
// workflow triggers and dispatches each match to the Executor. Errors here are
// logged but never returned — returning an error would cause NATS to redeliver,
// and Temporal's deterministic WorkflowID already handles dedup on retry.
func (c *Consumer) handle(event *events.DomainEvent) error {
	if event == nil {
		return nil
	}
	// Skip dispatch entirely when Temporal is unavailable so we don't burn
	// DB lookups for events we can't act on. Events will be lost (NATS won't
	// redeliver after the message is ack'd by the queue group) — operators
	// who care about replay should re-enable Temporal before going to prod.
	if !c.executor.Enabled() {
		return nil
	}

	workflows, err := c.store.FindMatchingWorkflows(event)
	if err != nil {
		log.Printf("[automation consumer] find matching workflows: %v", err)
		return nil
	}
	if len(workflows) == 0 {
		return nil
	}

	ctx := context.Background()
	for _, wf := range workflows {
		runID, err := c.executor.StartWorkflow(ctx, wf, event)
		if err != nil {
			log.Printf("[automation consumer] start workflow %s: %v", wf.ID, err)
			continue
		}
		log.Printf("[automation consumer] started workflow=%s run=%s trigger=%s entity=%s",
			wf.ID, runID, event.Subject, event.EntityID)
	}
	return nil
}
