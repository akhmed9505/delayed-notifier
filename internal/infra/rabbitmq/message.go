// Package rabbitmq provides infrastructure logic for RabbitMQ message consumption and retry handling.
package rabbitmq

import "time"

// NotificationMessage represents the data structure of a notification message processed through RabbitMQ.
type NotificationMessage struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Recipient string    `json:"recipient"`
	Channel   string    `json:"channel"`
	SendAt    time.Time `json:"send_at"`
	Attempt   int       `json:"attempt"`
}
