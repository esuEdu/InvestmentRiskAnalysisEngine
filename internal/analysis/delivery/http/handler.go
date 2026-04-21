package http

import (
	"net/http"
	"strconv"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
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

// ── POST /api/v1/analyses ─────────────────────────────────────────────────────
type CreateRequest struct {
	PortfolioID string  `json:"portfolio_id" binding:"required,uuid"`
	Benchmark   *string `json:"benchmark"`
	Period      string  `json:"period" binding:"required"` // e.g. "1y", "6m"
}

func (h *AnalysisHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pID := uuid.MustParse(req.PortfolioID) // safe: binding:"uuid" already validated

	result, err := h.useCase.ExecuteCreate(c.Request.Context(), pID, req.Benchmark, req.Period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue analysis"})
		return
	}

	// 202 Accepted signals the job is queued, not yet completed.
	c.JSON(http.StatusAccepted, result)
}

// ── GET /api/v1/analyses/:id ──────────────────────────────────────────────────

func (h *AnalysisHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis id"})
		return
	}

	result, err := h.useCase.ExecuteGet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analysis not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── GET /api/v1/analyses ──────────────────────────────────────────────────────

func (h *AnalysisHandler) List(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be a non-negative integer"})
		return
	}

	var statusFilter *domain.Status
	if s := c.Query("status"); s != "" {
		v := domain.Status(s)
		statusFilter = &v
	}

	results, err := h.useCase.ExecuteList(c.Request.Context(), int32(limit), int32(offset), statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list analyses"})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, []domain.AnalysisRequest{})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ── PUT /api/v1/analyses/:id ──────────────────────────────────────────────────

type UpdateRequest struct {
	Status domain.Status `json:"status" binding:"required"`
}

func (h *AnalysisHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analysis id"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.ExecuteUpdate(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update analysis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "status": req.Status})
}
