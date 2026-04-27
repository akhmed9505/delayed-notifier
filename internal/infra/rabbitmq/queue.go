package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type QueueConfig struct {
	Exchange   string
	Queue      string
	DLX        string
	DLQ        string
	RoutingKey string
}

func SetupQueues(ch *amqp091.Channel, cfg QueueConfig) error {
	if err := ch.ExchangeDeclare(cfg.Exchange, "x-delayed-message", true, false, false, false, amqp091.Table{
		"x-delayed-type": "direct",
	}); err != nil {
		return fmt.Errorf("exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(cfg.DLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("dlx: %w", err)
	}

	_, err := ch.QueueDeclare(cfg.DLQ, true, false, false, false, amqp091.Table{})
	if err != nil {
		return fmt.Errorf("dlq: %w", err)
	}

	_, err = ch.QueueDeclare(cfg.Queue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    cfg.DLX,
		"x-dead-letter-routing-key": cfg.DLQ,
	})
	if err != nil {
		return fmt.Errorf("queue: %w", err)
	}

	if err := ch.QueueBind(cfg.Queue, cfg.RoutingKey, cfg.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind: %w", err)
	}

	if err := ch.QueueBind(cfg.DLQ, cfg.DLQ, cfg.DLX, false, nil); err != nil {
		return fmt.Errorf("dlq bind: %w", err)
	}

	return nil
}
