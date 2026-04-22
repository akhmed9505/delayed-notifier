package middleware

import (
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

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
