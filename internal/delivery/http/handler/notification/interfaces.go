package notification

import (
	"context"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error)
	GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error
}

