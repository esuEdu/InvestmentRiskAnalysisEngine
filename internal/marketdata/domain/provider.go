package domain

import "context"

type Provider interface {
	FetchDailyPrices(ctx context.Context, ticker string) ([]PricePoint, error)
}
