// Package logger provides functionality for initializing and configuring the application's logging system.
package logger

import (
	"github.com/wb-go/wbf/zlog"

	"github.com/akhmed9505/delayed-notifier/internal/config"
)

// Init configures the global logger based on the provided application configuration.
// It initializes the logging backend and sets the log level if specified.
func Init(cfg *config.Config) {
	zlog.Init()

	if cfg.Logging.Level != "" {
		_ = zlog.SetLevel(cfg.Logging.Level)
	}
}
