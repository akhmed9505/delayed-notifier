package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/logger"
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
	log           *logger.ZerologAdapter
}

func NewNotificationHandler(
	svc NotificationStatusUpdater,
	email Mailer,
	telegram Mailer,
	log *logger.ZerologAdapter,
) *NotificationHandler {
	return &NotificationHandler{
		statusUpdater: svc,
		email:         email,
		telegram:      telegram,
		log:           log,
	}
}

func (h *NotificationHandler) Handle(ctx context.Context, body []byte) error {
	var msg struct {
		ID          uuid.UUID `json:"id"`
		Message     string    `json:"message"`
		Destination string    `json:"destination"`
		Channel     string    `json:"channel"`
		SendAt      time.Time `json:"send_at"`
	}

	if err := json.Unmarshal(body, &msg); err != nil {
		h.log.Error("unmarshal failed", "err", err, "body", string(body))
		return fmt.Errorf("unmarshal error: %w", err)
	}

	status, err := h.statusUpdater.Status(ctx, msg.ID)
	if err != nil {
		h.log.Error("status check failed", "id", msg.ID, "err", err)
		return fmt.Errorf("status check error: %w", err)
	}

	if status == "canceled" {
		h.log.Info("notification canceled, skipping", "id", msg.ID)
		return nil
	}

	if !msg.SendAt.IsZero() {
		wait := time.Until(msg.SendAt)
		if wait > 0 {
			h.log.Info("delayed execution", "id", msg.ID, "wait", wait)

			timer := time.NewTimer(wait)
			defer timer.Stop()

			select {
			case <-timer.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	var sender Mailer

	switch msg.Channel {
	case "email":
		sender = h.email
	case "telegram":
		sender = h.telegram
	default:
		h.log.Error("unknown channel", "id", msg.ID, "channel", msg.Channel)
		return fmt.Errorf("unknown channel: %s", msg.Channel)
	}

	if sender == nil {
		h.log.Error("nil sender", "id", msg.ID, "channel", msg.Channel)
		return fmt.Errorf("sender not initialized: %s", msg.Channel)
	}

	if err := sender.Send(ctx, msg.Message, msg.Destination); err != nil {
		h.log.Error("send failed", "id", msg.ID, "err", err)

		_ = h.statusUpdater.UpdateStatus(ctx, msg.ID, "failed")

		return fmt.Errorf("send failed: %w", err)
	}

	if err := h.statusUpdater.UpdateStatus(ctx, msg.ID, "sent"); err != nil {
		h.log.Error("status update failed", "id", msg.ID, "err", err)
		return fmt.Errorf("status update failed: %w", err)
	}

	h.log.Info("notification sent successfully", "id", msg.ID)
	return nil
}
