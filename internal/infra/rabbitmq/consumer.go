// Package rabbitmq provides infrastructure logic for RabbitMQ message consumption and retry handling.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// Handler defines the interface for processing notification messages.
type Handler interface {
	Handle(ctx context.Context, msg NotificationMessage) error
}

// RetryConfig holds the configuration for exponential backoff retry strategy.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	Multiplier  float64
	MaxDelay    time.Duration
	JitterPct   float64
}

// Consumer manages the RabbitMQ consumption process, including automatic retries.
type Consumer struct {
	ch         *amqp091.Channel
	queue      string
	exchange   string
	routingKey string
	retry      RetryConfig
	handler    Handler
	rnd        *rand.Rand
}

// NewConsumer initializes a new Consumer instance.
func NewConsumer(ch *amqp091.Channel, queue, exchange, routingKey string, retry RetryConfig, handler Handler) *Consumer {
	return &Consumer{
		ch:         ch,
		queue:      queue,
		exchange:   exchange,
		routingKey: routingKey,
		retry:      retry,
		handler:    handler,
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Start begins consuming messages from the queue and processing them using the provided handler.
// It handles retry logic with exponential backoff if the handler returns an error.
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

			err := c.handler.Handle(ctx, msg)
			if err == nil {
				_ = d.Ack(false)
				continue
			}

			// Calculate next attempt
			attempt := msg.Attempt
			if h := headerInt(d.Headers, "x-attempt"); h > attempt {
				attempt = h
			}

			nextAttempt := attempt + 1

			// Check if we reached the maximum number of attempts
			if c.retry.MaxAttempts > 0 && nextAttempt > c.retry.MaxAttempts {
				fmt.Printf("[retry] GIVE UP id=%s attempts=%d error=%v\n",
					msg.ID, nextAttempt-1, err)

				_ = d.Nack(false, false)
				continue
			}

			delay := c.retryDelay(nextAttempt)

			fmt.Printf("[retry] id=%s attempt=%d delay=%s error=%v\n",
				msg.ID, nextAttempt, delay, err)

			msg.Attempt = nextAttempt

			retryBody, mErr := json.Marshal(msg)
			if mErr != nil {
				_ = d.Nack(false, false)
				continue
			}

			// Prepare headers for the delayed retry
			headers := amqp091.Table{
				"x-delay":   delay.Milliseconds(),
				"x-attempt": nextAttempt,
			}

			// Re-publish the message to the retry exchange/queue
			if err := c.ch.PublishWithContext(ctx, c.exchange, c.routingKey, false, false, amqp091.Publishing{
				ContentType: "application/json",
				Body:        retryBody,
				Headers:     headers,
			}); err != nil {
				_ = d.Nack(false, true)
				continue
			}

			_ = d.Ack(false)
		}
	}
}

// headerInt is a helper to safely extract integer values from RabbitMQ headers.
func headerInt(headers amqp091.Table, key string) int {
	if headers == nil {
		return 0
	}

	switch v := headers[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	default:
		return 0
	}
}

// retryDelay calculates the delay duration based on exponential backoff and jitter.
func (c *Consumer) retryDelay(attempt int) time.Duration {
	base := c.retry.BaseDelay
	if base <= 0 {
		base = time.Second
	}

	mult := c.retry.Multiplier
	if mult <= 0 {
		mult = 2
	}

	maxDelay := c.retry.MaxDelay
	if maxDelay <= 0 {
		maxDelay = time.Minute
	}

	raw := float64(base) * math.Pow(mult, float64(attempt-1))
	delay := time.Duration(raw)

	if delay > maxDelay {
		delay = maxDelay
	}

	jp := c.retry.JitterPct
	if jp <= 0 {
		return delay
	}

	if jp > 1 {
		jp = 1
	}

	f := 1 + (c.rnd.Float64()*2-1)*jp
	final := time.Duration(float64(delay) * f)

	if final < 0 {
		return 0
	}

	return final
}
