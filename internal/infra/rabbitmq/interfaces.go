//go:generate mockgen -source=interfaces.go -destination=mocks.go -package=rabbitmq
package rabbitmq

import "github.com/rabbitmq/amqp091-go"

type ChannelProvider interface {
	GetChannel() (*amqp091.Channel, error)
}
