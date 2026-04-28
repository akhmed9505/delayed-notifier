//go:generate mockgen -source=interfaces.go -destination=mocks.go -package=notification

// Package notification provides service-level interfaces for managing notification operations.
package notification

import (
	"context"

	"github.com/google/uuid"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

// Repository defines the persistence layer interface for notification operations.
type Repository interface {
	// Create persists a new notification and returns its unique identifier.
	Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error)

	// GetStatusByID retrieves the status of a specific notification by its ID.
	GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)

	// UpdateStatus updates the status of an existing notification.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error
}

// Publisher defines the interface for publishing notification messages to the message broker.
type Publisher interface {
	// Publish sends the notification message to the queue.
	Publish(ctx context.Context, notification domain.Notification) error
}

// Cache defines the interface for caching notification statuses.
type Cache interface {
	// SetStatus saves the notification status in the cache.
	SetStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error

	// GetStatus retrieves the notification status from the cache.
	GetStatus(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)
}
