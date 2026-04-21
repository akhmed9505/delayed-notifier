package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
)

const (
	Exchange   = "notification-exchange"
	RoutingKey = "notification.send"
)

type Publisher struct {
	pub *rabbitmq.Publisher
}

type Message struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Recipient string    `json:"recipient"`
	Channel   string    `json:"channel"`
	SendAt    time.Time `json:"send_at"`
}

func NewPublisher(client *rabbitmq.RabbitClient) (*Publisher, error) {
	ch, err := client.GetChannel()
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		Exchange,
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-delayed-type": "direct",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	pub := rabbitmq.NewPublisher(client, Exchange, "application/json")

	return &Publisher{pub: pub}, nil
}

func (p *Publisher) Publish(ctx context.Context, notification domain.Notification) error {
	msg := Message{
		ID:        notification.ID.String(),
		Message:   notification.Message,
		Recipient: notification.Recipient,
		Channel:   string(notification.Channel),
		SendAt:    notification.SendAt,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	delay := time.Until(msg.SendAt)
	if delay < 0 {
		delay = 0
	}

	opts := []rabbitmq.PublishOption{
		func(pub *amqp091.Publishing) {
			if pub.Headers == nil {
				pub.Headers = amqp091.Table{}
			}
			pub.Headers["x-delay"] = delay.Milliseconds()
		},
	}

	if err := p.pub.Publish(ctx, body, RoutingKey, opts...); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}
