package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type ChannelProvider interface {
	GetChannel() (*amqp.Channel, error)
}
