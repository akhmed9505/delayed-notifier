// Package redis provides infrastructure logic for Redis client management.
package redis

import (
	"fmt"

	wbfredis "github.com/wb-go/wbf/redis"

	"github.com/akhmed9505/delayed-notifier/internal/config"
)

// Client wraps the underlying Redis client to provide application-specific functionality.
type Client struct {
	*wbfredis.Client
}

// New creates and returns a new Redis client instance based on the provided configuration.
func New(cfg *config.Redis) *Client {
	if cfg == nil {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &Client{Client: wbfredis.New(addr, cfg.Password, cfg.DB)}
}
