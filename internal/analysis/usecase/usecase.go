package usecase

import (
	"context"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/google/uuid"
)

type UseCase struct {
	repo  domain.Repository
	queue domain.Queue
}

func New(r domain.Repository, q domain.Queue) *UseCase {
	return &UseCase{repo: r, queue: q}
}

func (u *UseCase) ExecuteCreate(ctx context.Context, assets []domain.Asset, benchmark *string, period string) (domain.AnalysisRequest, error) {
	analysis := domain.AnalysisRequest{
		ID:        uuid.New(),
		Assets:    assets,
		Benchmark: benchmark,
		Period:    period,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}

	created, err := u.repo.Create(ctx, analysis)
	if err != nil {
		return domain.AnalysisRequest{}, err
	}

	if err := u.queue.PublishAnalysisJob(&created); err != nil {
		return domain.AnalysisRequest{}, err
	}

	return created, nil
}

func (u *UseCase) ExecuteGet(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
	return u.repo.Get(ctx, id)
}

func (u *UseCase) ExecuteUpdate(ctx context.Context, id uuid.UUID, status domain.Status) error {
	return u.repo.UpdateStatus(ctx, id, status)
}

func (u *UseCase) ExecuteList(ctx context.Context, limit, offset int32, status *domain.Status) ([]domain.AnalysisRequest, error) {
	return u.repo.List(ctx, limit, offset, (*string)(status))
}

func (u *UseCase) ExecuteGetResult(ctx context.Context, id uuid.UUID) (domain.AnalysisResult, error) {
	return u.repo.GetAnalysisResult(ctx, id)
}
