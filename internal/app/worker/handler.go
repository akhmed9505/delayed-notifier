package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/infra/rabbitmq"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"
)

type Mailer interface {
	Send(ctx context.Context, message, destination string) error
}

type NotificationStatusUpdater interface {
	UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error
	Status(ctx context.Context, noteID uuid.UUID) (string, error)
}

type NotificationHandler struct {
	statusUpdater NotificationStatusUpdater
	email         Mailer
	telegram      Mailer
	log           zlog.Zerolog
}

func NewNotificationHandler(
	svc NotificationStatusUpdater,
	email Mailer,
	telegram Mailer,
	log zlog.Zerolog,
) *NotificationHandler {
	return &NotificationHandler{
		statusUpdater: svc,
		email:         email,
		telegram:      telegram,
		log:           log,
	}
}

func (h *NotificationHandler) Handle(ctx context.Context, msg rabbitmq.NotificationMessage) error {
	id, err := uuid.Parse(msg.ID)
	if err != nil {
		h.log.Error().Str("id", msg.ID).Err(err).Msg("invalid notification id")
		return fmt.Errorf("invalid notification id: %w", err)
	}

	status, err := h.statusUpdater.Status(ctx, id)
	if err != nil {
		h.log.Error().Str("id", id.String()).Err(err).Msg("status check failed")
		return fmt.Errorf("status check error: %w", err)
	}

	if status != "pending" {
		h.log.Info().
			Str("id", id.String()).
			Str("status", status).
			Int("attempt", msg.Attempt).
			Msg("skip: not pending")
		return nil
	}

	var sender Mailer

	switch msg.Channel {
	case "email":
		sender = h.email
	case "telegram":
		sender = h.telegram
	default:
		h.log.Error().
			Str("id", id.String()).
			Str("channel", msg.Channel).
			Msg("unknown channel")

		_ = h.statusUpdater.UpdateStatus(ctx, id, "failed")
		return fmt.Errorf("unknown channel: %s", msg.Channel)
	}

	const maxRetries = 3

	var sendErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		sendErr = sender.Send(ctx, msg.Message, msg.Recipient)

		if sendErr == nil {
			h.log.Info().
				Str("id", id.String()).
				Int("attempt", attempt).
				Msg("send success")

			break
		}

		h.log.Error().
			Str("id", id.String()).
			Int("attempt", attempt).
			Err(sendErr).
			Msg("send attempt failed")

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}

	if sendErr != nil {
		_ = h.statusUpdater.UpdateStatus(ctx, id, "failed")

		return fmt.Errorf("send failed after %d attempts: %w", maxRetries, sendErr)
	}

	if err := h.statusUpdater.UpdateStatus(ctx, id, "sent"); err != nil {
		h.log.Error().Str("id", id.String()).Err(err).Msg("status update failed")
		return fmt.Errorf("status update failed: %w", err)
	}

	h.log.Info().
		Str("id", id.String()).
		Msg("notification sent successfully")

	return nil
}
