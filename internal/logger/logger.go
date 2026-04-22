package logger

import (
	"github.com/akhmed9505/delayed-notifier/internal/config"
	"github.com/wb-go/wbf/zlog"
)

func Init(cfg *config.Config) {
	zlog.Init()

	if cfg.Logging.Level != "" {
		_ = zlog.SetLevel(cfg.Logging.Level)
	}
}
