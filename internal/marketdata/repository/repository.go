package repository

import (
	"context"
	"errors"
	"time"

	sqlc "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/esuEdu/investment-risk-engine/internal/marketdata/domain"
	"github.com/jackc/pgx/v5"
)

type Repo struct {
	queries *sqlc.Queries
}

func New(q *sqlc.Queries) *Repo {
	return &Repo{queries: q}
}

func (r *Repo) GetPricesSince(ctx context.Context, ticker string, since time.Time) ([]domain.PricePoint, error) {
	rows, err := r.queries.GetHistoricalPricesSince(ctx, sqlc.GetHistoricalPricesSinceParams{
		Ticker:    ticker,
		PriceDate: toPgDate(since),
	})
	if err != nil {
		return nil, err
	}
	prices := make([]domain.PricePoint, 0, len(rows))
	for _, row := range rows {
		prices = append(prices, toPricePoint(row))
	}
	return prices, nil
}

func (r *Repo) GetLatestPriceDate(ctx context.Context, ticker string) (time.Time, error) {
	d, err := r.queries.GetLatestPriceDate(ctx, ticker)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return d.Time, nil
}

func (r *Repo) UpsertPrices(ctx context.Context, ticker string, prices []domain.PricePoint) error {
	for _, p := range prices {
		vol := p.Volume
		if err := r.queries.UpsertHistoricalPrice(ctx, sqlc.UpsertHistoricalPriceParams{
			Ticker:    ticker,
			PriceDate: toPgDate(p.Date),
			Open:      pgNumeric(p.Open),
			High:      pgNumeric(p.High),
			Low:       pgNumeric(p.Low),
			Close:     pgNumeric(p.Close),
			Volume:    &vol,
		}); err != nil {
			return err
		}
	}
	return nil
}
