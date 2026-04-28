// Package rabbitmq provides infrastructure logic for RabbitMQ client management and channel handling.
package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	wbrabbitmq "github.com/wb-go/wbf/rabbitmq"
)

// Client wraps the underlying RabbitMQ client to provide connection management.
type Client struct {
	inner *wbrabbitmq.RabbitClient
}

// NewClient initializes a new RabbitMQ client instance with the provided connection URL.
func NewClient(url string) (*Client, error) {
	inner, err := wbrabbitmq.NewClient(wbrabbitmq.ClientConfig{
		URL: url,
	})
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connect: %w", err)
	}

	return &Client{inner: inner}, nil
}

// GetChannel returns a new communication channel from the RabbitMQ connection.
func (c *Client) GetChannel() (*amqp091.Channel, error) {
	if c == nil || c.inner == nil {
		return nil, fmt.Errorf("rabbitmq client not initialized")
	}
	return c.inner.GetChannel()
}

// Close terminates the connection to the RabbitMQ broker.
func (c *Client) Close() error {
	if c == nil || c.inner == nil {
		return nil
	}
	return c.inner.Close()
}
