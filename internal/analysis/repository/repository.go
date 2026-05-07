package repository

import (
	"context"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func New(q *sqlc.Queries, pool *pgxpool.Pool) *Repo {
	return &Repo{queries: q, pool: pool}
}

func (r *Repo) Create(ctx context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.AnalysisRequest{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.CreateAnalysisRequest(ctx, sqlc.CreateAnalysisRequestParams{
		ID:        uuidToPg(req.ID),
		Status:    string(req.Status),
		Benchmark: req.Benchmark,
		Period:    req.Period,
	})
	if err != nil {
		return domain.AnalysisRequest{}, err
	}

	for _, a := range req.Assets {
		if err := qtx.InsertAnalysisRequestAsset(ctx, sqlc.InsertAnalysisRequestAssetParams{
			AnalysisRequestID: uuidToPg(req.ID),
			Ticker:            a.Ticker,
			Weight:            pgNumeric(a.Weight),
		}); err != nil {
			return domain.AnalysisRequest{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.AnalysisRequest{}, err
	}

	result := toDomain(row)
	result.Assets = req.Assets
	return result, nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
	row, err := r.queries.GetAnalysisRequest(ctx, uuidToPg(id))
	if err != nil {
		return domain.AnalysisRequest{}, err
	}

	req := toDomain(row)

	assets, err := r.GetAssets(ctx, id)
	if err != nil {
		return domain.AnalysisRequest{}, err
	}
	req.Assets = assets

	return req, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	return r.queries.UpdateAnalysisRequestStatus(ctx, sqlc.UpdateAnalysisRequestStatusParams{
		ID:     uuidToPg(id),
		Status: string(status),
	})
}

func (r *Repo) List(ctx context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error) {
	var s string
	if status != nil {
		s = *status
	}

	rows, err := r.queries.ListAnalysisRequests(ctx, sqlc.ListAnalysisRequestsParams{
		Limit:   limit,
		Offset:  offset,
		Column3: s,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.AnalysisRequest, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomain(row))
	}
	return result, nil
}

func (r *Repo) GetAssets(ctx context.Context, id uuid.UUID) ([]domain.Asset, error) {
	rows, err := r.queries.GetAnalysisRequestAssets(ctx, uuidToPg(id))
	if err != nil {
		return nil, err
	}

	assets := make([]domain.Asset, 0, len(rows))
	for _, row := range rows {
		assets = append(assets, domain.Asset{
			Ticker: row.Ticker,
			Weight: numericToFloat64(row.Weight),
		})
	}
	return assets, nil
}

func (r *Repo) GetAnalysisResult(ctx context.Context, id uuid.UUID) (domain.AnalysisResult, error) {
	row, err := r.queries.GetAnalysisResult(ctx, uuidToPg(id))
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	return resultToDomain(row), nil
}

func (r *Repo) CreateAnalysisResult(ctx context.Context, id uuid.UUID, res domain.AnalysisResult) error {
	return r.queries.InsertAnalysisResult(ctx, sqlc.InsertAnalysisResultParams{
		AnalysisRequestID:    uuidToPg(id),
		AnnualizedVolatility: pgNumeric(res.AnnualizedVolatility),
		SharpeRatio:          pgNumeric(res.SharpeRatio),
		Beta:                 pgNumericPtr(res.Beta),
		MaxDrawdown:          pgNumeric(res.MaxDrawdown),
		Var95:                pgNumeric(res.VaR95),
		ConcentrationScore:   pgNumeric(res.ConcentrationScore),
		RawMetricsJson:       res.RawMetricsJSON,
	})
}
