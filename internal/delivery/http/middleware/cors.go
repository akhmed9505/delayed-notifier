// Package middleware provides common HTTP middleware components for the application.
package middleware

import "github.com/wb-go/wbf/ginext"

// CORS returns a middleware that handles Cross-Origin Resource Sharing by setting
// appropriate headers and responding to OPTIONS requests.
func CORS() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.Writer.WriteHeader(200)
			return
		}

		c.Next()
	}
}
