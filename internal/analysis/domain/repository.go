package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, req AnalysisRequest) (AnalysisRequest, error)
	Get(ctx context.Context, id uuid.UUID) (AnalysisRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	List(ctx context.Context, limit, offset int32, status *string) ([]AnalysisRequest, error)
}
