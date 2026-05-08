package http

import (
	"strconv"

	"github.com/esuEdu/investment-risk-engine/internal/portfolio/domain"
	"github.com/esuEdu/investment-risk-engine/internal/portfolio/usecase"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/esuEdu/investment-risk-engine/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	uc        *usecase.UseCase
	jwtSecret string
}

func New(uc *usecase.UseCase, jwtSecret string) *Handler {
	return &Handler{uc: uc, jwtSecret: jwtSecret}
}

// ── POST /api/v1/auth/register ────────────────────────────────────────────────

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	token, user, err := h.uc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		logger.Log.Warnw("registration failed", "email", req.Email, "error", err)
		response.BadRequest(c, "email already registered")
		return
	}

	response.OK(c, "registered successfully", gin.H{
		"token": token,
		"user":  gin.H{"id": user.ID, "email": user.Email},
	})
}

// ── POST /api/v1/auth/login ───────────────────────────────────────────────────

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	token, user, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.BadRequest(c, "invalid credentials")
		return
	}

	response.OK(c, "logged in successfully", gin.H{
		"token": token,
		"user":  gin.H{"id": user.ID, "email": user.Email},
	})
}

// ── POST /api/v1/portfolios ───────────────────────────────────────────────────

type assetRequest struct {
	Ticker string  `json:"ticker" binding:"required"`
	Weight float64 `json:"weight" binding:"required,gt=0,lte=1"`
}

type createPortfolioRequest struct {
	Name   string         `json:"name"   binding:"required"`
	Assets []assetRequest `json:"assets" binding:"required,min=1,dive"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet(contextUserID).(uuid.UUID)

	var req createPortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	assets := make([]domain.PortfolioAsset, len(req.Assets))
	for i, a := range req.Assets {
		assets[i] = domain.PortfolioAsset{Ticker: a.Ticker, Weight: a.Weight}
	}

	p, err := h.uc.CreatePortfolio(c.Request.Context(), userID, req.Name, assets)
	if err != nil {
		logger.Log.Warnw("create portfolio failed", "user_id", userID, "error", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "portfolio created", p)
}

// ── GET /api/v1/portfolios ────────────────────────────────────────────────────

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet(contextUserID).(uuid.UUID)

	portfolios, err := h.uc.ListPortfolios(c.Request.Context(), userID)
	if err != nil {
		logger.Log.Errorw("list portfolios failed", "user_id", userID, "error", err)
		response.InternalError(c, "failed to list portfolios")
		return
	}

	if portfolios == nil {
		portfolios = []domain.Portfolio{}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response.OKList(c, "portfolios listed", portfolios, response.Meta{
		Limit:  limit,
		Offset: offset,
		Count:  len(portfolios),
	})
}

// ── GET /api/v1/portfolios/:id ────────────────────────────────────────────────

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet(contextUserID).(uuid.UUID)

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid portfolio id")
		return
	}

	p, err := h.uc.GetPortfolio(c.Request.Context(), userID, portfolioID)
	if err != nil {
		response.NotFound(c, "portfolio not found")
		return
	}

	response.OK(c, "portfolio retrieved", p)
}

// ── PUT /api/v1/portfolios/:id ────────────────────────────────────────────────

type updatePortfolioRequest struct {
	Name   string         `json:"name"   binding:"required"`
	Assets []assetRequest `json:"assets" binding:"required,min=1,dive"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet(contextUserID).(uuid.UUID)

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid portfolio id")
		return
	}

	var req updatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	assets := make([]domain.PortfolioAsset, len(req.Assets))
	for i, a := range req.Assets {
		assets[i] = domain.PortfolioAsset{Ticker: a.Ticker, Weight: a.Weight}
	}

	p, err := h.uc.UpdatePortfolio(c.Request.Context(), userID, portfolioID, req.Name, assets)
	if err != nil {
		logger.Log.Warnw("update portfolio failed", "portfolio_id", portfolioID, "error", err)
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "portfolio updated", p)
}

// ── DELETE /api/v1/portfolios/:id ─────────────────────────────────────────────

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet(contextUserID).(uuid.UUID)

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid portfolio id")
		return
	}

	if err := h.uc.DeletePortfolio(c.Request.Context(), userID, portfolioID); err != nil {
		response.NotFound(c, "portfolio not found")
		return
	}

	response.OK(c, "portfolio deleted", gin.H{"id": portfolioID})
}

// ── POST /api/v1/portfolios/:id/analyses ─────────────────────────────────────

type triggerAnalysisRequest struct {
	Benchmark *string `json:"benchmark"`
	Period    string  `json:"period" binding:"required"`
}

func (h *Handler) TriggerAnalysis(c *gin.Context) {
	userID := c.MustGet(contextUserID).(uuid.UUID)

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid portfolio id")
		return
	}

	var req triggerAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.uc.TriggerAnalysis(c.Request.Context(), userID, portfolioID, req.Benchmark, req.Period)
	if err != nil {
		logger.Log.Errorw("trigger analysis failed", "portfolio_id", portfolioID, "error", err)
		response.InternalError(c, "failed to trigger analysis")
		return
	}

	response.Accepted(c, "analysis queued", result)
}
