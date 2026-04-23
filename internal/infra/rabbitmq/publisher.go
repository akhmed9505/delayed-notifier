package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	client ChannelProvider
	cfg    QueueConfig
}

func NewPublisher(client ChannelProvider, cfg QueueConfig) *Publisher {
	return &Publisher{
		client: client,
		cfg:    cfg,
	}
}

func (p *Publisher) Publish(ctx context.Context, msg NotificationMessage) error {
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

	headers := amqp.Table{
		"x-delay": delay.Milliseconds(),
	}

	return ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		p.cfg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Headers:      headers,
			DeliveryMode: amqp.Persistent,
		},
	)
}
