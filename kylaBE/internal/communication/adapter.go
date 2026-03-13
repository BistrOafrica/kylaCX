package communication

import (
	"context"
	"fmt"
)

// ChannelAdapter defines the interface for sending messages via any channel.
type ChannelAdapter interface {
	Send(ctx context.Context, conv *Conversation, msg *Message) error
	Channel() string
}

// AdapterRegistry dispatches outbound messages to the correct channel adapter.
type AdapterRegistry struct {
	adapters map[string]ChannelAdapter
}

// NewAdapterRegistry constructs an AdapterRegistry from a list of adapters.
func NewAdapterRegistry(adapters ...ChannelAdapter) *AdapterRegistry {
	registry := &AdapterRegistry{
		adapters: make(map[string]ChannelAdapter),
	}
	for _, a := range adapters {
		registry.adapters[a.Channel()] = a
	}
	return registry
}

// Dispatch routes the message to the appropriate adapter based on conv.Channel.
func (r *AdapterRegistry) Dispatch(ctx context.Context, conv *Conversation, msg *Message) error {
	adapter, ok := r.adapters[conv.Channel]
	if !ok {
		return fmt.Errorf("no adapter registered for channel %s", conv.Channel)
	}
	return adapter.Send(ctx, conv, msg)
}
