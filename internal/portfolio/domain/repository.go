package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, u User) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)

	CreatePortfolio(ctx context.Context, p Portfolio) (Portfolio, error)
	GetPortfolio(ctx context.Context, id uuid.UUID) (Portfolio, error)
	ListPortfolios(ctx context.Context, userID uuid.UUID) ([]Portfolio, error)
	UpdatePortfolio(ctx context.Context, id uuid.UUID, name string, assets []PortfolioAsset) (Portfolio, error)
	DeletePortfolio(ctx context.Context, id uuid.UUID) error
}
