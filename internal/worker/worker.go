package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/riskmetrics"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"github.com/google/uuid"
)

// MarketDataClient is the interface the Worker uses to fetch historical prices.
// Implemented by HTTPMarketDataClient; can be mocked in tests.
type MarketDataClient interface {
	GetPrices(ctx context.Context, ticker, period string) ([]riskmetrics.PricePoint, error)
}

// AnalysisRepository is the subset of domain.Repository the Worker needs.
type AnalysisRepository interface {
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error
	GetAssets(ctx context.Context, id uuid.UUID) ([]domain.Asset, error)
	CreateAnalysisResult(ctx context.Context, id uuid.UUID, result domain.AnalysisResult) error
}

type Handler struct {
	repo       AnalysisRepository
	mdClient   MarketDataClient
	calculator riskmetrics.Calculator
}

func New(repo AnalysisRepository, md MarketDataClient, calc riskmetrics.Calculator) *Handler {
	return &Handler{repo: repo, mdClient: md, calculator: calc}
}

func (h *Handler) Handle(ctx context.Context, req *domain.AnalysisRequest) error {
	logger.Log.Infow("processing analysis job", "analysis_id", req.ID, "period", req.Period)

	if err := h.repo.UpdateStatus(ctx, req.ID, domain.StatusProcessing); err != nil {
		return fmt.Errorf("update status processing: %w", err)
	}

	assets, err := h.repo.GetAssets(ctx, req.ID)
	if err != nil {
		h.markFailed(ctx, req.ID, "get assets")
		return fmt.Errorf("get assets: %w", err)
	}

	rmAssets := make([]riskmetrics.Asset, 0, len(assets))
	for _, a := range assets {
		prices, err := h.mdClient.GetPrices(ctx, a.Ticker, req.Period)
		if err != nil {
			h.markFailed(ctx, req.ID, "fetch prices for "+a.Ticker)
			return fmt.Errorf("fetch prices for %s: %w", a.Ticker, err)
		}
		rmAssets = append(rmAssets, riskmetrics.Asset{
			Ticker: a.Ticker,
			Prices: prices,
			Weight: a.Weight,
		})
	}

	var benchmarkPrices []riskmetrics.PricePoint
	if req.Benchmark != nil {
		benchmarkPrices, err = h.mdClient.GetPrices(ctx, *req.Benchmark, req.Period)
		if err != nil {
			h.markFailed(ctx, req.ID, "fetch benchmark prices")
			return fmt.Errorf("fetch benchmark prices: %w", err)
		}
	}

	metricsResult, err := h.calculator.Compute(riskmetrics.MetricsInput{
		Assets:          rmAssets,
		BenchmarkPrices: benchmarkPrices,
		RiskFreeRate:    0.05,
	})
	if err != nil {
		h.markFailed(ctx, req.ID, "compute metrics")
		return fmt.Errorf("compute metrics: %w", err)
	}

	rawJSON, _ := json.Marshal(metricsResult)

	domainResult := domain.AnalysisResult{
		AnnualizedVolatility: metricsResult.AnnualizedVolatility,
		SharpeRatio:          metricsResult.SharpeRatio,
		Beta:                 metricsResult.Beta,
		MaxDrawdown:          metricsResult.MaxDrawdown,
		VaR95:                metricsResult.VaR95,
		ConcentrationScore:   metricsResult.ConcentrationScore,
		RawMetricsJSON:       rawJSON,
	}

	if err := h.repo.CreateAnalysisResult(ctx, req.ID, domainResult); err != nil {
		h.markFailed(ctx, req.ID, "save result")
		return fmt.Errorf("save result: %w", err)
	}

	if err := h.repo.UpdateStatus(ctx, req.ID, domain.StatusCompleted); err != nil {
		logger.Log.Errorw("failed to mark analysis completed", "analysis_id", req.ID, "error", err)
		return err
	}

	logger.Log.Infow("analysis job completed", "analysis_id", req.ID)
	return nil
}

func (h *Handler) markFailed(ctx context.Context, id uuid.UUID, step string) {
	if err := h.repo.UpdateStatus(ctx, id, domain.StatusFailed); err != nil {
		logger.Log.Errorw("failed to mark analysis as failed", "analysis_id", id, "step", step, "error", err)
	}
}
