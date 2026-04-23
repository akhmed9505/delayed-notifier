package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Handle(ctx context.Context, msg NotificationMessage) error
}

type Consumer struct {
	ch      *amqp.Channel
	queue   string
	handler Handler
}

func NewConsumer(ch *amqp.Channel, queue string, handler Handler) *Consumer {
	return &Consumer{
		ch:      ch,
		queue:   queue,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.ch.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case d, ok := <-deliveries:
			if !ok {
				return nil
			}

			var msg NotificationMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				_ = d.Nack(false, false)
				continue
			}

			if err := c.handler.Handle(ctx, msg); err != nil {
				_ = d.Nack(false, true)
				continue
			}

			_ = d.Ack(false)
		}
	}
}
