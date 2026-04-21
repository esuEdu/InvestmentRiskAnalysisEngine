package http

import (
	"strconv"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/usecase"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/esuEdu/investment-risk-engine/pkg/response"
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
		response.BadRequest(c, err.Error())
		return
	}

	pID := uuid.MustParse(req.PortfolioID)

	result, err := h.useCase.ExecuteCreate(c.Request.Context(), pID, req.Benchmark, req.Period)
	if err != nil {
		logger.Log.Errorw("failed to create analysis",
			"portfolio_id", req.PortfolioID,
			"period", req.Period,
			"error", err,
		)
		response.InternalError(c, "failed to queue analysis")
		return
	}

	logger.Log.Infow("analysis queued",
		"analysis_id", result.ID,
		"portfolio_id", req.PortfolioID,
		"period", result.Period,
	)

	response.Accepted(c, "analysis queued successfully", result)
}

// ── GET /api/v1/analyses/:id ──────────────────────────────────────────────────

func (h *AnalysisHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid analysis id")
		return
	}

	result, err := h.useCase.ExecuteGet(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "analysis not found")
		return
	}

	response.OK(c, "analysis retrieved", result)
}

// ── GET /api/v1/analyses ──────────────────────────────────────────────────────

func (h *AnalysisHandler) List(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		response.BadRequest(c, "limit must be a positive integer")
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		response.BadRequest(c, "offset must be a non-negative integer")
		return
	}

	var statusFilter *domain.Status
	if s := c.Query("status"); s != "" {
		v := domain.Status(s)
		statusFilter = &v
	}

	results, err := h.useCase.ExecuteList(c.Request.Context(), int32(limit), int32(offset), statusFilter)
	if err != nil {
		logger.Log.Errorw("failed to list analyses",
			"limit", limit,
			"offset", offset,
			"status_filter", statusFilter,
			"error", err,
		)
		response.InternalError(c, "failed to list analyses")
		return
	}

	if results == nil {
		results = []domain.AnalysisRequest{}
	}

	response.OKList(c, "analyses listed", results, response.Meta{
		Limit:  limit,
		Offset: offset,
		Count:  len(results),
	})
}

// ── PUT /api/v1/analyses/:id ──────────────────────────────────────────────────

type UpdateRequest struct {
	Status domain.Status `json:"status" binding:"required"`
}

func (h *AnalysisHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid analysis id")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.useCase.ExecuteUpdate(c.Request.Context(), id, req.Status); err != nil {
		logger.Log.Errorw("failed to update analysis status",
			"analysis_id", id,
			"status", req.Status,
			"error", err,
		)
		response.InternalError(c, "failed to update analysis")
		return
	}

	logger.Log.Infow("analysis status updated",
		"analysis_id", id,
		"status", req.Status,
	)

	response.OK(c, "analysis updated", gin.H{"id": id, "status": req.Status})
}
