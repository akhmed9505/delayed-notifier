//go:generate mockgen -source=interfaces.go -destination=mocks.go -package=rabbitmq

// Package rabbitmq provides infrastructure logic for RabbitMQ client management and channel handling.
package rabbitmq

import "github.com/rabbitmq/amqp091-go"

// ChannelProvider defines an interface for components capable of providing a RabbitMQ channel.
type ChannelProvider interface {
	// GetChannel returns a new RabbitMQ channel.
	GetChannel() (*amqp091.Channel, error)
}
