package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/akhmed9505/delayed-notifier/internal/app"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		log.Printf("failed to create app: %v", err)
		return
	}

	zlog.Logger.Info().Msg("starting notification worker")
	go func() {
		if err := a.Worker.Run(ctx); err != nil && err != context.Canceled {
			zlog.Logger.Error().Err(err).Msg("worker stopped with error")
		} else {
			zlog.Logger.Info().Msg("worker stopped gracefully")
		}
	}()

	zlog.Logger.Info().Msg("starting http server")
	if err := a.Run(ctx); err != nil {
		zlog.Logger.Error().Err(err).Msg("app stopped with error")
	}
}
