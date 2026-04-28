// Package sender provides infrastructure logic for sending notifications via various channels.
package sender

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/akhmed9505/delayed-notifier/internal/config"
)

// TelegramChannel implements the notification sender interface for sending messages via Telegram.
type TelegramChannel struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramChannel initializes a new Telegram bot instance with the provided configuration.
func NewTelegramChannel(cfg *config.Telegram) (*TelegramChannel, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("telegram bot init: %w", err)
	}

	return &TelegramChannel{bot: bot}, nil
}

// Send transmits a message to the specified Telegram chat ID.
func (t *TelegramChannel) Send(ctx context.Context, message, recipient string) error {
	chatID, err := strconv.ParseInt(recipient, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat_id: %w", err)
	}

	msg := tgbotapi.NewMessage(chatID, message)

	done := make(chan error, 1)
	go func() {
		_, err := t.bot.Send(msg)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case err := <-done:
		if err != nil {
			return fmt.Errorf("telegram send: %w", err)
		}
	}

	return nil
}
