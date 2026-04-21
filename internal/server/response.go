package server

import (
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/esuEdu/investment-risk-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

// OK delegates to the shared response package.
func OK(c *gin.Context, message string, data any) {
	response.OK(c, message, data)
}

// BadRequest logs the error then returns a 400 response.
func BadRequest(c *gin.Context, message string, err any) {
	logger.Log.Warnw("Bad Request",
		"path", c.Request.URL.Path,
		"error", err,
	)
	response.BadRequest(c, message)
}

// InternalError logs the error then returns a 500 response.
func InternalError(c *gin.Context, message string, err any) {
	logger.Log.Errorw("Internal Server Error",
		"path", c.Request.URL.Path,
		"error", err,
	)
	response.InternalError(c, message)
}
