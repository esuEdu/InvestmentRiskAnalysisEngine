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
