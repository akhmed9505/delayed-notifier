//go:generate mockgen -source=interfaces.go -destination=mocks.go -package=notification

// Package notification defines the service interface required by the HTTP handlers.
package notification

import (
	"context"

	"github.com/google/uuid"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

// Service defines the methods required for managing notifications in the service layer.
type Service interface {
	// Create initiates a new notification creation.
	Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error)

	// GetStatusByID retrieves the status of a notification by its unique identifier.
	GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)

	// UpdateStatus changes the status of a specific notification.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error
}
