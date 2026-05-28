package nats

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

// Client wraps a NATS connection.
type Client struct {
	conn *nats.Conn
}

// Connect establishes a NATS connection. url defaults to nats.DefaultURL if empty.
func Connect(url string) (*Client, error) {
	if url == "" {
		url = nats.DefaultURL
	}
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats: connect to %s: %w", url, err)
	}
	log.Printf("NATS connected to %s", url)
	return &Client{conn: conn}, nil
}

// Close drains and closes the connection gracefully.
func (c *Client) Close() {
	if c.conn != nil {
		if err := c.conn.Drain(); err != nil {
			log.Printf("NATS drain: %v", err)
		}
	}
}

// Conn exposes the underlying *nats.Conn for advanced use.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}
