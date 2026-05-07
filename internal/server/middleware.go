package server

import (
	"github.com/esuEdu/investment-risk-engine/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func ZapLogger() gin.HandlerFunc {
	return middleware.ZapLogger()
}
