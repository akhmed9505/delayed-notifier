package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	wbrabbitmq "github.com/wb-go/wbf/rabbitmq"
)

type Client struct {
	inner *wbrabbitmq.RabbitClient
}

func NewClient(url string) (*Client, error) {
	inner, err := wbrabbitmq.NewClient(wbrabbitmq.ClientConfig{
		URL: url,
	})
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connect: %w", err)
	}

	return &Client{inner: inner}, nil
}

func (c *Client) GetChannel() (*amqp091.Channel, error) {
	if c == nil || c.inner == nil {
		return nil, fmt.Errorf("rabbitmq client not initialized")
	}
	return c.inner.GetChannel()
}

func (c *Client) Close() error {
	if c == nil || c.inner == nil {
		return nil
	}
	return c.inner.Close()
}
