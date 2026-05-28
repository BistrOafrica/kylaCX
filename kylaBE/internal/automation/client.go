package automation

import (
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
)

// ClientConfig holds the Temporal client connection parameters.
type ClientConfig struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

// NewTemporalClient dials the Temporal frontend and returns a connected client.
// Returns (nil, nil) when HostPort is empty so the binary can boot without Temporal
// in environments where automation is disabled.
func NewTemporalClient(cfg ClientConfig) (client.Client, error) {
	if cfg.HostPort == "" {
		log.Println("temporal: TEMPORAL_HOST_PORT not set; automation engine disabled")
		return nil, nil
	}
	ns := cfg.Namespace
	if ns == "" {
		ns = "default"
	}
	c, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: ns,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal dial %s: %w", cfg.HostPort, err)
	}
	log.Printf("temporal: connected to %s (namespace=%s, task_queue=%s)", cfg.HostPort, ns, cfg.TaskQueue)
	return c, nil
}
