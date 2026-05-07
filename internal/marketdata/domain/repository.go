package domain

import (
	"context"
	"time"
)

type Repository interface {
	GetPricesSince(ctx context.Context, ticker string, since time.Time) ([]PricePoint, error)
	GetLatestPriceDate(ctx context.Context, ticker string) (time.Time, error)
	UpsertPrices(ctx context.Context, ticker string, prices []PricePoint) error
}
