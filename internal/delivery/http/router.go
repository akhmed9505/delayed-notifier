// Package http defines the HTTP routing and server configuration for the application.
package http

import (
	"github.com/wb-go/wbf/ginext"

	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/handler/notification"
	"github.com/akhmed9505/delayed-notifier/internal/delivery/http/middleware"
)

// NewRouter creates a new ginext engine, configures global middleware, and registers notification routes.
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
