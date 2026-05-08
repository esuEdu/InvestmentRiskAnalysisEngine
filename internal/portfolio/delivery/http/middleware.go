package http

import (
	"strings"

	"github.com/esuEdu/investment-risk-engine/pkg/auth"
	"github.com/esuEdu/investment-risk-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

const contextUserID = "user_id"

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.BadRequest(c, "missing or malformed Authorization header")
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(strings.TrimPrefix(header, "Bearer "), jwtSecret)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(contextUserID, claims.UserID)
		c.Next()
	}
}
