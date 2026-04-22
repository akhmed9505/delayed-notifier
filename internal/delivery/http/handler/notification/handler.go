package notification

import (
	"net/http"

	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/helpers"
	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/response"
	"github.com/akhmed9505/delayed-notifier/internal/domain"
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

	sendAt, err := helpers.ParseSendAt(req.SendAt)
	if err != nil {
		response.BadRequest(c, err.Error())
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

func (h *Handler) GetStatus(c *ginext.Context) {
	id, err := helpers.ParseUUIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
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
	id, err := helpers.ParseUUIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), id, domain.Canceled); err != nil {
		response.InternalError(c, errCancelFailed)
		return
	}

	c.Status(http.StatusNoContent)
}
