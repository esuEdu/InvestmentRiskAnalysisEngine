package http

import (
	"errors"
	"strings"

	"github.com/esuEdu/investment-risk-engine/internal/marketdata/usecase"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/esuEdu/investment-risk-engine/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	useCase *usecase.UseCase
}

func New(uc *usecase.UseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) GetPrices(c *gin.Context) {
	ticker := c.Query("ticker")
	period := c.Query("period")

	if ticker == "" || period == "" {
		response.BadRequest(c, "ticker and period query params are required")
		return
	}

	prices, err := h.useCase.GetPrices(c.Request.Context(), ticker, period)
	if err != nil {
		logger.Log.Errorw("failed to get prices", "ticker", ticker, "period", period, "error", err)

		// Surface quota/rate-limit errors as 503 so the caller can retry later.
		if isQuotaError(err) {
			c.JSON(503, gin.H{"error": err.Error()})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "prices retrieved", prices)
}

func isQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := errors.Unwrap(err)
	s := err.Error()
	if msg != nil {
		s = msg.Error()
	}
	return strings.Contains(s, "quota") ||
		strings.Contains(s, "rate limit") ||
		strings.Contains(s, "not active")
}
