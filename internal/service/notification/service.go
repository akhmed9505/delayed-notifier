package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/akhmed9505/delayed-notifier/internal/infra/redis"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error)
	GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error
}

type Publisher interface {
	Publish(ctx context.Context, notification domain.Notification) error
}

type Cache interface {
	SetStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error
	GetStatus(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error)
}

type Service struct {
	repository Repository
	publisher  Publisher
	cache      Cache
}

func New(repository Repository, publisher Publisher, cache Cache) *Service {
	return &Service{
		repository: repository,
		publisher:  publisher,
		cache:      cache,
	}
}

func (s *Service) Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error) {
	const op = "notification.service.Create"

	id, err := s.repository.Create(ctx, notification)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	notification.ID = id

	_ = s.cache.SetStatus(ctx, id, domain.Pending)

	if err := s.publisher.Publish(ctx, notification); err != nil {
		return uuid.Nil, fmt.Errorf("%s: publish notification: %w", op, err)
	}

	return id, nil
}

func (s *Service) GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error) {
	const op = "notification.service.GetStatusByID"

	status, err := s.cache.GetStatus(ctx, id)
	if err == nil {
		return status, nil
	}
	if !errors.Is(err, redis.ErrCacheMiss) {
		return "", fmt.Errorf("%s: get status from cache: %w", op, err)
	}

	status, err = s.repository.GetStatusByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	_ = s.cache.SetStatus(ctx, id, status)

	return status, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	const op = "notification.service.UpdateStatus"

	if err := s.repository.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_ = s.cache.SetStatus(ctx, id, status)

	return nil
}
