package http

import (
	"net/http"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnalysisHandler struct {
	useCase *usecase.UseCase
}

func New(u *usecase.UseCase) *AnalysisHandler {
	return &AnalysisHandler{useCase: u}
}

type CreateRequest struct {
	PortfolioID string  `json:"portfolio_id" binding:"required,uuid"`
	Benchmark   *string `json:"benchmark"`
	Period      string  `json:"period" binding:"required"` // e.g., "1y"
}

func (h *AnalysisHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pID := uuid.MustParse(req.PortfolioID)

	result, err := h.useCase.ExecuteCreate(c.Request.Context(), pID, req.Benchmark, req.Period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue analysis"})
		return
	}

	// 202 Accepted is standard for asynchronous background jobs
	c.JSON(http.StatusAccepted, result)
}
