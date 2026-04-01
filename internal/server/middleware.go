package server

import (
	"time"

	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/gin-gonic/gin"
)

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logFields := []interface{}{
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"ip", c.ClientIP(),
			"latency", latency.String(),
			"user-agent", c.Request.UserAgent(),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				logger.Log.Errorw("Request Error", "error", e)
			}
		} else {
			if status >= 500 {
				logger.Log.Errorw("Server Error", logFields...)
			} else if status >= 400 {
				logger.Log.Warnw("Client Error", logFields...)
			} else {
				logger.Log.Infow("Request Processed", logFields...)
			}
		}
	}
}
