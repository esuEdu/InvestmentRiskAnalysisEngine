package repository

import (
	"context"

	"github.com/esuEdu/investment-risk-engine/internal/portfolio/domain"
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

func (r *Repo) CreateUser(ctx context.Context, u domain.User) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           uuidToPg(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
	})
	if err != nil {
		return domain.User{}, err
	}
	return userToDomain(row), nil
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return userToDomain(row), nil
}

func (r *Repo) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, uuidToPg(id))
	if err != nil {
		return domain.User{}, err
	}
	return userToDomain(row), nil
}

func (r *Repo) CreatePortfolio(ctx context.Context, p domain.Portfolio) (domain.Portfolio, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Portfolio{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.CreatePortfolio(ctx, sqlc.CreatePortfolioParams{
		ID:     uuidToPg(p.ID),
		UserID: uuidToPg(p.UserID),
		Name:   p.Name,
	})
	if err != nil {
		return domain.Portfolio{}, err
	}

	for _, a := range p.Assets {
		if err := qtx.InsertPortfolioAsset(ctx, sqlc.InsertPortfolioAssetParams{
			PortfolioID: uuidToPg(p.ID),
			Ticker:      a.Ticker,
			Weight:      pgNumeric(a.Weight),
		}); err != nil {
			return domain.Portfolio{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Portfolio{}, err
	}

	result := portfolioToDomain(row)
	result.Assets = p.Assets
	return result, nil
}

func (r *Repo) GetPortfolio(ctx context.Context, id uuid.UUID) (domain.Portfolio, error) {
	row, err := r.queries.GetPortfolio(ctx, uuidToPg(id))
	if err != nil {
		return domain.Portfolio{}, err
	}

	p := portfolioToDomain(row)

	assetRows, err := r.queries.GetPortfolioAssets(ctx, uuidToPg(id))
	if err != nil {
		return domain.Portfolio{}, err
	}
	p.Assets = assetsToDomain(assetRows)

	return p, nil
}

func (r *Repo) ListPortfolios(ctx context.Context, userID uuid.UUID) ([]domain.Portfolio, error) {
	rows, err := r.queries.ListPortfoliosByUser(ctx, uuidToPg(userID))
	if err != nil {
		return nil, err
	}

	portfolios := make([]domain.Portfolio, 0, len(rows))
	for _, row := range rows {
		p := portfolioToDomain(row)
		assetRows, err := r.queries.GetPortfolioAssets(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		p.Assets = assetsToDomain(assetRows)
		portfolios = append(portfolios, p)
	}
	return portfolios, nil
}

func (r *Repo) UpdatePortfolio(ctx context.Context, id uuid.UUID, name string, assets []domain.PortfolioAsset) (domain.Portfolio, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Portfolio{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.UpdatePortfolioName(ctx, sqlc.UpdatePortfolioNameParams{
		ID:   uuidToPg(id),
		Name: name,
	})
	if err != nil {
		return domain.Portfolio{}, err
	}

	if err := qtx.DeletePortfolioAssets(ctx, uuidToPg(id)); err != nil {
		return domain.Portfolio{}, err
	}

	for _, a := range assets {
		if err := qtx.InsertPortfolioAsset(ctx, sqlc.InsertPortfolioAssetParams{
			PortfolioID: uuidToPg(id),
			Ticker:      a.Ticker,
			Weight:      pgNumeric(a.Weight),
		}); err != nil {
			return domain.Portfolio{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Portfolio{}, err
	}

	result := portfolioToDomain(row)
	result.Assets = assets
	return result, nil
}

func (r *Repo) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeletePortfolio(ctx, uuidToPg(id))
}
