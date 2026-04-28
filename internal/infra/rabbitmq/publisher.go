// Package rabbitmq provides infrastructure logic for RabbitMQ message publishing and consumption.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// Publisher is responsible for sending notification messages to the RabbitMQ exchange.
type Publisher struct {
	client ChannelProvider
	cfg    QueueConfig
}

// NewPublisher initializes a new Publisher instance with the given client and configuration.
func NewPublisher(client ChannelProvider, cfg QueueConfig) *Publisher {
	return &Publisher{
		client: client,
		cfg:    cfg,
	}
}

// Publish sends a notification message to the configured exchange with a delay based on the SendAt field.
func (p *Publisher) Publish(ctx context.Context, msg NotificationMessage) error {
	msg.Attempt = 0
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	ch, err := p.client.GetChannel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	defer ch.Close()

	delay := time.Until(msg.SendAt)
	if delay < 0 {
		delay = 0
	}

	headers := amqp091.Table{
		"x-delay":   delay.Milliseconds(),
		"x-attempt": 0,
	}

	return ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		p.cfg.RoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Headers:      headers,
			DeliveryMode: amqp091.Persistent,
		},
	)
}
