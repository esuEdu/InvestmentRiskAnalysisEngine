package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body is the standard envelope for every API response.
//
//	Success:  { "message": "...", "data": ... }
//	List:     { "message": "...", "data": [...], "meta": { ... } }
//	Error:    { "error": "..." }
type Body struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// Meta holds pagination information returned on list endpoints.
type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Body{Message: message, Data: data})
}

func OKList(c *gin.Context, message string, data any, meta Meta) {
	c.JSON(http.StatusOK, Body{Message: message, Data: data, Meta: &meta})
}

func Accepted(c *gin.Context, message string, data any) {
	c.JSON(http.StatusAccepted, Body{Message: message, Data: data})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, Body{Error: err})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusNotFound, Body{Error: err})
}

func InternalError(c *gin.Context, err string) {
	c.JSON(http.StatusInternalServerError, Body{Error: err})
}
