package redis

import (
	"context"
	"errors"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

var ErrCacheMiss = errors.New("cache miss")

type Cache struct {
	client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) SetStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	return c.client.Set(ctx, id.String(), string(status))
}

func (c *Cache) SetStatusWithRetry(ctx context.Context, id uuid.UUID, status domain.NotificationStatus, strategy retry.Strategy) error {
	return c.client.SetWithRetry(ctx, strategy, id.String(), string(status))
}

func (c *Cache) GetStatus(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error) {
	status, err := c.client.Get(ctx, id.String())
	if err != nil {
		return "", ErrCacheMiss
	}

	if status == "" {
		return "", ErrCacheMiss
	}

	return domain.NotificationStatus(status), nil
}

func (c *Cache) Delete(ctx context.Context, id uuid.UUID) error {
	return c.client.Del(ctx, id.String())
}

