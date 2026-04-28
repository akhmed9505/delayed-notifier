// Package domain defines the core business entities and types for the delayed-notifier application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationChannel represents the delivery method for a notification.
type NotificationChannel string

const (
	// Email represents the email notification channel.
	Email NotificationChannel = "email"

	// Telegram represents the Telegram notification channel.
	Telegram NotificationChannel = "telegram"
)

// NotificationStatus represents the current state of a notification.
type NotificationStatus string

const (
	// Pending represents a notification waiting to be processed.
	Pending NotificationStatus = "pending"

	// Sent represents a notification that has been successfully delivered.
	Sent NotificationStatus = "sent"

	// Canceled represents a notification that has been canceled by the user.
	Canceled NotificationStatus = "canceled"

	// Failed represents a notification that could not be delivered.
	Failed NotificationStatus = "failed"
)

// Notification represents the core business entity for a scheduled notification.
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
