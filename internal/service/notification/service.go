package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error)
	GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error) {
	const op = "notification.service.Create"

	id, err := s.repository.Create(ctx, notification)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Service) GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error) {
	const op = "notification.service.GetStatusByID"

	status, err := s.repository.GetStatusByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return status, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	const op = "notification.service.UpdateStatus"

	if err := s.repository.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
