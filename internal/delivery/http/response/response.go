// Package response provides standardized structures and helper functions for HTTP error responses.
package response

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

// errorResponse defines the structure for JSON error responses.
type errorResponse struct {
	Error string `json:"error"`
}

// JSONError sends a standardized JSON error response with the given status code and message.
func JSONError(c *ginext.Context, status int, msg string) {
	c.JSON(status, errorResponse{Error: msg})
}

// BadRequest sends a 400 Bad Request error response.
func BadRequest(c *ginext.Context, msg string) {
	JSONError(c, http.StatusBadRequest, msg)
}

// NotFound sends a 404 Not Found error response.
func NotFound(c *ginext.Context, msg string) {
	JSONError(c, http.StatusNotFound, msg)
}

// InternalError sends a 500 Internal Server Error response.
func InternalError(c *ginext.Context, msg string) {
	JSONError(c, http.StatusInternalServerError, msg)
}
