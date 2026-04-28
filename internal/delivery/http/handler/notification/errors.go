// Package notification provides common error messages used within the notification delivery handler.
package notification

const (
	// errInvalidJSON is returned when the request body fails to parse.
	errInvalidJSON = "invalid json"

	// errCreateFailed is returned when the notification creation process fails in the service layer.
	errCreateFailed = "failed to create notification"

	// errInvalidSendAt is returned when the provided time is in the past.
	errInvalidSendAt = "send_at must be in the future"

	// errStatusFailed is returned when the status retrieval fails.
	errStatusFailed = "failed to get status"

	// errCancelFailed is returned when the notification cancellation process fails.
	errCancelFailed = "failed to cancel notification"

	// errNotFound is returned when the requested notification does not exist.
	errNotFound = "notification not found"
)
