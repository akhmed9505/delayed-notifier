package notification

import (
	"errors"
	"net/http"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/response"
	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *ginext.Context) {
	var req createRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, errInvalidJSON)
		return
	}

	sendAt, err := time.Parse(time.RFC3339, req.SendAt)
	if err != nil {
		response.BadRequest(c, errInvalidSendAt)
		return
	}

	channel, err := parseChannel(req.Channel)
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

func (h *Handler) GetStatus(c *ginext.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, errInvalidID)
		return
	}

	status, err := h.svc.GetStatusByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, errStatusFailed)
		return
	}

	c.JSON(http.StatusOK, statusResponse{
		ID:     id.String(),
		Status: string(status),
	})
}

func (h *Handler) Cancel(c *ginext.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, errInvalidID)
		return
	}

	err = h.svc.UpdateStatus(c.Request.Context(), id, domain.Canceled)
	if err != nil {
		response.InternalError(c, errCancelFailed)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseChannel(s string) (domain.NotificationChannel, error) {
	switch s {
	case string(domain.Email):
		return domain.Email, nil
	case string(domain.Telegram):
		return domain.Telegram, nil
	default:
		return "", errors.New("invalid channel")
	}
}
