package notification

import (
	"context"
	"errors"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/akhmed9505/delayed-notifier/internal/infra/redis"
	"github.com/google/uuid"
)

var ErrCacheMiss = errors.New("cache miss")

type StatusCache struct {
	client *redis.Client
}

func NewStatusCache(client *redis.Client) *StatusCache {
	return &StatusCache{client: client}
}

func (c *StatusCache) SetStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	if c == nil || c.client == nil {
		return errors.New("status cache: not initialized")
	}
	return c.client.Set(ctx, id.String(), string(status))
}

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
