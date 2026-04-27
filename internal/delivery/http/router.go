package http

import (
	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/handler/notification"
	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/middleware"
	"github.com/wb-go/wbf/ginext"
)

func NewRouter(handler *notification.Handler) *ginext.Engine {
	r := ginext.New("")

	r.Use(ginext.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	if handler == nil {
		return r
	}

	r.POST("/notify", handler.Create)
	r.GET("/notify/:id", handler.GetStatus)
	r.DELETE("/notify/:id", handler.Cancel)

	return r
}
