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
	msk := time.FixedZone("MSK", 3*60*60)

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, msk); err == nil {
		return t, nil
	}

	return time.Time{}, ErrInvalidSendAt
}

func ParseUUIDParam(c *ginext.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)

	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, ErrInvalidID
	}

	return id, nil
}
