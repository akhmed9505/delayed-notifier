// Package notification handles HTTP requests for creating, checking, and canceling notifications.
package notification

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"

	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/helpers"
	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/response"
	"github.com/akhmed9505/delayed-notifier/internal/domain"
	notifyrepo "github.com/akhmed9505/delayed-notifier/internal/repository/notification"
)

// Handler manages notification-related HTTP endpoints.
type Handler struct {
	svc Service
}

// New initializes a new Handler with the provided Service implementation.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles the request to create a new notification.
// It parses the JSON body, validates input, and delegates the creation to the service layer.
func (h *Handler) Create(c *ginext.Context) {
	var req createRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errInvalidJSON)
		return
	}

	sendAt, err := helpers.ParseSendAt(req.SendAt)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if sendAt.UTC().Before(time.Now().UTC()) {
		response.BadRequest(c, errInvalidSendAt)
		return
	}

	channel, err := helpers.ParseChannel(req.Channel)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	notification := domain.Notification{
		Message:   req.Message,
		Channel:   channel,
		Recipient: req.Recipient,
		SendAt:    sendAt,
		Status:    domain.Pending,
	}

	id, err := h.svc.Create(c.Request.Context(), notification)
	if err != nil {
		response.InternalError(c, errCreateFailed)
		return
	}

	c.JSON(http.StatusCreated, createResponse{ID: id.String()})
}

// GetStatus retrieves the status of a notification by its ID.
func (h *Handler) GetStatus(c *ginext.Context) {
	id, err := helpers.ParseUUIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	status, err := h.svc.GetStatusByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, notifyrepo.ErrNotificationNotFound) {
			response.NotFound(c, errNotFound)
			return
		}
		response.InternalError(c, errStatusFailed)
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		ID:     id.String(),
		Status: string(status),
	})
}

// Cancel updates the status of a notification to "canceled".
func (h *Handler) Cancel(c *ginext.Context) {
	id, err := helpers.ParseUUIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), id, domain.Canceled); err != nil {
		if errors.Is(err, notifyrepo.ErrNotificationNotFound) {
			response.NotFound(c, errNotFound)
			return
		}
		response.InternalError(c, errCancelFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification canceled",
	})
}
