package repository

import (
	"context"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/google/uuid"
)

type Repo struct {
	queries *sqlc.Queries
}

func New(q *sqlc.Queries) *Repo {
	return &Repo{queries: q}
}

func (r *Repo) Create(ctx context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
	row, err := r.queries.CreateAnalysisRequest(ctx, sqlc.CreateAnalysisRequestParams{
		ID:        uuidToPg(req.ID),
		Status:    string(req.Status),
		Benchmark: req.Benchmark,
		Period:    req.Period,
	})
	if err != nil {
		return domain.AnalysisRequest{}, err
	}
	return toDomain(row), nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
	row, err := r.queries.GetAnalysisRequest(ctx, uuidToPg(id))
	if err != nil {
		return domain.AnalysisRequest{}, err
	}
	return toDomain(row), nil
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
