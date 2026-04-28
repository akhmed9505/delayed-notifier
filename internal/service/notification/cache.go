// Package notification provides service-level logic for handling notifications, including caching mechanisms.
package notification

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/akhmed9505/delayed-notifier/internal/infra/redis"
)

// ErrCacheMiss is returned when the requested notification status is not found in the cache.
var ErrCacheMiss = errors.New("cache miss")

// StatusCache provides caching capabilities for notification statuses using Redis.
type StatusCache struct {
	client *redis.Client
}

// NewStatusCache initializes a new StatusCache with the provided Redis client.
func NewStatusCache(client *redis.Client) *StatusCache {
	return &StatusCache{client: client}
}

// SetStatus stores the notification status in the cache.
func (c *StatusCache) SetStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	if c == nil || c.client == nil {
		return errors.New("status cache: not initialized")
	}
	return c.client.Set(ctx, id.String(), string(status))
}

// GetStatus retrieves the notification status from the cache.
// Returns ErrCacheMiss if the status is not found.
func (c *StatusCache) GetStatus(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error) {
	if c == nil || c.client == nil {
		return "", errors.New("status cache: not initialized")
	}
	status, err := c.client.Get(ctx, id.String())
	if err != nil {
		return "", err
	}
	if status == "" {
		return "", ErrCacheMiss
	}
	return domain.NotificationStatus(status), nil
}
