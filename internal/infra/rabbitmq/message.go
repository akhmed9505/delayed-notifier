package rabbitmq

import "time"

type NotificationMessage struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Recipient string    `json:"recipient"`
	Channel   string    `json:"channel"`
	SendAt    time.Time `json:"send_at"`
}
