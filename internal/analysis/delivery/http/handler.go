package http

import (
	"math"
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

type AssetRequest struct {
	Ticker string  `json:"ticker" binding:"required"`
	Weight float64 `json:"weight" binding:"required,gt=0,lte=1"`
}

type CreateRequest struct {
	Assets    []AssetRequest `json:"assets"    binding:"required,min=1,dive"`
	Benchmark *string        `json:"benchmark"`
	Period    string         `json:"period"    binding:"required"`
}

func (h *AnalysisHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var total float64
	for _, a := range req.Assets {
		total += a.Weight
	}
	if math.Abs(total-1.0) > 1e-4 {
		response.BadRequest(c, "asset weights must sum to 1.0")
		return
	}

	assets := make([]domain.Asset, len(req.Assets))
	for i, a := range req.Assets {
		assets[i] = domain.Asset{Ticker: a.Ticker, Weight: a.Weight}
	}

	result, err := h.useCase.ExecuteCreate(c.Request.Context(), assets, req.Benchmark, req.Period)
	if err != nil {
		logger.Log.Errorw("failed to create analysis", "period", req.Period, "error", err)
		response.InternalError(c, "failed to queue analysis")
		return
	}

	logger.Log.Infow("analysis queued", "analysis_id", result.ID, "period", result.Period)
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
		logger.Log.Errorw("failed to list analyses", "limit", limit, "offset", offset, "error", err)
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
		logger.Log.Errorw("failed to update analysis status", "analysis_id", id, "status", req.Status, "error", err)
		response.InternalError(c, "failed to update analysis")
		return
	}

	logger.Log.Infow("analysis status updated", "analysis_id", id, "status", req.Status)
	response.OK(c, "analysis updated", gin.H{"id": id, "status": req.Status})
}

// ── GET /api/v1/analyses/:id/results ─────────────────────────────────────────

func (h *AnalysisHandler) GetResult(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid analysis id")
		return
	}

	result, err := h.useCase.ExecuteGetResult(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "results not found")
		return
	}

	response.OK(c, "results retrieved", result)
}
