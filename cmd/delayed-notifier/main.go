// Package main implements the entry point for the delayed-notifier service.
package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/wb-go/wbf/zlog"

	"github.com/akhmed9505/delayed-notifier/internal/app"
)

// main initializes the application, configures signal handling for graceful shutdown,
// starts the background notification worker, and executes the HTTP server.
func main() {
	// Setup context to handle SIGINT and SIGTERM signals for graceful termination.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize the application and its dependencies.
	a, err := app.New(ctx)
	if err != nil {
		log.Printf("failed to create app: %v", err)
		return
	}

	// Start the notification worker in a background goroutine.
	zlog.Logger.Info().Msg("starting notification worker")
	go func() {
		if err := a.Worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			zlog.Logger.Error().Err(err).Msg("worker stopped with error")
		} else {
			zlog.Logger.Info().Msg("worker stopped gracefully")
		}
	}()

	// Start the HTTP server. This operation is blocking.
	zlog.Logger.Info().Msg("starting http server")
	if err := a.Run(ctx); err != nil {
		zlog.Logger.Error().Err(err).Msg("app stopped with error")
	}
}
