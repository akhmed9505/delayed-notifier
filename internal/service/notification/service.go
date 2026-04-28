// Package notification provides service-level logic for handling notifications, coordinating between storage, caching, and message publishing.
package notification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

// Service orchestrates the notification workflow, managing creation, status retrieval, and status updates by interacting with the repository, cache, and publisher.
type Service struct {
	repository Repository
	publisher  Publisher
	cache      Cache
}

// New initializes a new Service with the required repository, publisher, and cache dependencies.
func New(repository Repository, publisher Publisher, cache Cache) *Service {
	return &Service{
		repository: repository,
		publisher:  publisher,
		cache:      cache,
	}
}

// Create persists a new notification, caches its initial status, and publishes it for processing.
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

// GetStatusByID retrieves the current status of a notification, attempting to fetch it from the cache first and falling back to the database.
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

// UpdateStatus updates the status of a notification in both the persistent store and the cache.
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
