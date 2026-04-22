package helpers

import (
	"errors"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
)

var (
	ErrInvalidChannel = errors.New("invalid channel")
	ErrInvalidSendAt  = errors.New("invalid send_at, expected RFC3339")
	ErrInvalidID      = errors.New("invalid id")
)

func ParseChannel(s string) (domain.NotificationChannel, error) {
	switch s {
	case string(domain.Email):
		return domain.Email, nil
	case string(domain.Telegram):
		return domain.Telegram, nil
	default:
		return "", ErrInvalidChannel
	}
}

func ParseSendAt(s string) (time.Time, error) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.ParseInLocation(time.RFC3339, s, loc)
	if err != nil {
		return time.Time{}, ErrInvalidSendAt
	}

	return t, nil
}

func ParseUUIDParam(c *ginext.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)

	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, ErrInvalidID
	}

	return id, nil
}
