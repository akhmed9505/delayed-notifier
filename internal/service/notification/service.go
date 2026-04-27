package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"
)

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
	now := time.Now()
	notification.CreatedAt = now
	notification.UpdatedAt = now

	id, err := s.repository.Create(ctx, notification)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("repository create failed")
		return uuid.Nil, fmt.Errorf("create notification: %w", err)
	}

	notification.ID = id

	if err := s.cache.SetStatus(ctx, id, domain.Pending); err != nil {
		zlog.Logger.Warn().
			Err(err).
			Str("id", id.String()).
			Msg("cache set status failed")
	}

	zlog.Logger.Info().
		Str("id", id.String()).
		Msg("notification created")

	if err := s.publisher.Publish(ctx, notification); err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("id", id.String()).
			Msg("publish failed")

		return uuid.Nil, fmt.Errorf("publish notification: %w", err)
	}

	zlog.Logger.Info().
		Str("id", id.String()).
		Msg("notification published")

	return id, nil
}

func (s *Service) GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error) {
	status, err := s.cache.GetStatus(ctx, id)
	if err == nil {
		zlog.Logger.Info().
			Str("id", id.String()).
			Msg("status found in cache")

		return status, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		zlog.Logger.Error().
			Err(err).
			Str("id", id.String()).
			Msg("cache error")

		return "", fmt.Errorf("cache get status: %w", err)
	}

	zlog.Logger.Warn().
		Str("id", id.String()).
		Msg("cache miss, fallback to repository")

	status, err = s.repository.GetStatusByID(ctx, id)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("id", id.String()).
			Msg("repository lookup failed")

		return "", fmt.Errorf("get status: %w", err)
	}

	_ = s.cache.SetStatus(ctx, id, status)

	return status, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	if err := s.repository.UpdateStatus(ctx, id, status); err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("id", id.String()).
			Msg("update status failed")

		return fmt.Errorf("update status: %w", err)
	}

	_ = s.cache.SetStatus(ctx, id, status)

	zlog.Logger.Info().
		Str("id", id.String()).
		Str("status", string(status)).
		Msg("status updated")

	return nil
}
