package server

import (
	"net/http"

	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func JSON(c *gin.Context, status int, message string, data any, err any) {
	c.JSON(status, Response{
		Message: message,
		Data:    data,
		Error:   err,
	})
}

func OK(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, message, data, nil)
}

func Created(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusCreated, message, data, nil)
}

func BadRequest(c *gin.Context, message string, err interface{}) {

	logger.Log.Warnw("Bad Request",
		"path", c.Request.URL.Path,
		"error", err,
	)

	JSON(c, http.StatusBadRequest, message, nil, err)
}

func InternalError(c *gin.Context, message string, err interface{}) {

	logger.Log.Errorw("Internal Server Error",
		"path", c.Request.URL.Path,
		"error", err,
	)

	JSON(c, http.StatusInternalServerError, message, nil, err)
}
