// Package middleware provides common HTTP middleware components for the application.
package middleware

import (
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// Logger returns a middleware that logs details about incoming HTTP requests,
// including the HTTP method, request path, status code, latency in milliseconds, and client IP.
func Logger() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		start := time.Now()

		c.Next()

		zlog.Logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Str("ip", c.ClientIP()).
			Msg("http request")
	}
}
