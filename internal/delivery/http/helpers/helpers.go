// Package helpers provides utility functions for HTTP request parsing and validation.
package helpers

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

var (
	// ErrInvalidChannel indicates that the provided notification channel is not supported.
	ErrInvalidChannel = errors.New("invalid channel")

	// ErrInvalidSendAt indicates that the provided timestamp format is incorrect.
	ErrInvalidSendAt = errors.New("invalid send_at, expected RFC3339")

	// ErrInvalidID indicates that the provided parameter is not a valid UUID.
	ErrInvalidID = errors.New("invalid id")
)

// ParseChannel converts a string representation of a channel into a domain.NotificationChannel.
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

// ParseSendAt parses a date string into a time.Time object.
// It supports RFC3339 format and a specific local format (2006-01-02T15:04:05) in the MSK timezone.
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

// ParseUUIDParam extracts a UUID parameter from the Gin request context.
func ParseUUIDParam(c *ginext.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)

	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, ErrInvalidID
	}

	return id, nil
}
