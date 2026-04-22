package response

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

type errorResponse struct {
	Error string `json:"error"`
}

func JSONError(c *ginext.Context, status int, msg string) {
	c.JSON(status, errorResponse{Error: msg})
}

func BadRequest(c *ginext.Context, msg string) {
	JSONError(c, http.StatusBadRequest, msg)
}

func InternalError(c *ginext.Context, msg string) {
	JSONError(c, http.StatusInternalServerError, msg)
}

