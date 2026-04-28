// Package notification defines the Data Transfer Objects (DTOs) for the notification HTTP API.
package notification

// createRequest represents the schema for the incoming notification creation request.
type createRequest struct {
	Message   string `json:"message" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Recipient string `json:"recipient" binding:"required"`
	SendAt    string `json:"send_at" binding:"required"`
}

// createResponse represents the response containing the ID of the created notification.
type createResponse struct {
	ID string `json:"id"`
}

// statusResponse represents the response containing the status of a specific notification.
type statusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
