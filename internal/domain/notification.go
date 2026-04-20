package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationChannel string

const (
	Email    NotificationChannel = "email"
	Telegram NotificationChannel = "telegram"
)

type NotificationStatus string

const (
	Pending  NotificationStatus = "pending"
	Sent     NotificationStatus = "sent"
	Canceled NotificationStatus = "canceled"
	Failed   NotificationStatus = "failed"
)

type Notification struct {
	ID        uuid.UUID
	Message   string
	Channel   NotificationChannel
	Recipient string
	SendAt    time.Time
	Status    NotificationStatus
	Retries   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
