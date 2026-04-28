// Package http defines the HTTP server configuration and routing for the application.
package http

import (
	"net/http"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/config"
	"github.com/wb-go/wbf/ginext"
)

// NewServer creates and returns a configured http.Server instance with the given address and router.
func NewServer(addr string, router *ginext.Engine, cfg config.HTTPServer) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		IdleTimeout:       60 * time.Second,
	}
}
