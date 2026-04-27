package redis

import (
	"fmt"

	"github.com/akhmed9505/delayed-notifier/internal/config"
	wbfredis "github.com/wb-go/wbf/redis"
)

type Client struct {
	*wbfredis.Client
}

func New(cfg *config.Redis) *Client {
	if cfg == nil {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &Client{Client: wbfredis.New(addr, cfg.Password, cfg.DB)}
}
