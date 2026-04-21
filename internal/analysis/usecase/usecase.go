package usecase

import (
	"context"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/google/uuid"
)

type UseCase struct {
	repo domain.Repository
}

func New(r domain.Repository) *UseCase {
	return &UseCase{repo: r}
}

func (u *UseCase) ExecuteCreate(ctx context.Context, portfolioID uuid.UUID, benchmark *string, period string) (domain.AnalysisRequest, error) {
	analysis := domain.AnalysisRequest{
		ID:        uuid.New(),
		Benchmark: benchmark,
		Period:    period,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}

	created, err := u.repo.Create(ctx, analysis)
	if err != nil {
		return domain.AnalysisRequest{}, err
	}

	return created, nil
}

func (u *UseCase) ExecuteGet(ctx context.Context, AnalysisID uuid.UUID) (domain.AnalysisRequest, error) {
	analysis, err := u.repo.Get(ctx, AnalysisID)
	if err != nil {
		return domain.AnalysisRequest{}, err
	}

	return analysis, err
}

func (u *UseCase) ExecuteUpdate(ctx context.Context, AnalysisID uuid.UUID, status domain.Status) error {
	if err := u.repo.UpdateStatus(ctx, AnalysisID, status); err != nil {
		return err
	}

	return nil
}

func (u *UseCase) ExecuteList(ctx context.Context, limit, offset int32, status *domain.Status) ([]domain.AnalysisRequest, error) {
	listed, err := u.repo.List(ctx, limit, offset, (*string)(status))
	if err != nil {
		return nil, err
	}

	return listed, nil
}
