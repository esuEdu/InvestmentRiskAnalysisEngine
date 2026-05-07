package usecase

import (
	"context"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/marketdata/domain"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
)

type UseCase struct {
	repo     domain.Repository
	provider domain.Provider
}

func New(repo domain.Repository, provider domain.Provider) *UseCase {
	return &UseCase{repo: repo, provider: provider}
}

// GetPrices returns historical prices for ticker for the given period.
// It checks the DB cache first; if data is stale it fetches from AlphaVantage.
// If the external fetch fails (e.g. quota exceeded) it falls back to whatever
// is already cached so the worker can still proceed with older data.
func (u *UseCase) GetPrices(ctx context.Context, ticker, period string) ([]domain.PricePoint, error) {
	cutoff := periodCutoff(period)

	latest, err := u.repo.GetLatestPriceDate(ctx, ticker)
	if err != nil {
		return nil, err
	}

	if !isFresh(latest) {
		prices, fetchErr := u.provider.FetchDailyPrices(ctx, ticker)
		if fetchErr != nil {
			logger.Log.Warnw("AlphaVantage fetch failed, trying cache fallback",
				"ticker", ticker, "error", fetchErr)

			// Fall back to whatever is cached — return an error only if nothing is cached.
			cached, cacheErr := u.repo.GetPricesSince(ctx, ticker, cutoff)
			if cacheErr != nil || len(cached) == 0 {
				return nil, fetchErr // surface the original fetch error
			}
			logger.Log.Infow("Serving stale cached prices", "ticker", ticker, "points", len(cached))
			return cached, nil
		}

		if err := u.repo.UpsertPrices(ctx, ticker, prices); err != nil {
			return nil, err
		}
	}

	return u.repo.GetPricesSince(ctx, ticker, cutoff)
}

// isFresh returns true when latestDate is the most recent completed trading day.
func isFresh(latestDate time.Time) bool {
	if latestDate.IsZero() {
		return false
	}
	return !latestDate.Before(lastTradingDay(time.Now().UTC()))
}

// lastTradingDay returns the most recent completed trading day (skips weekends).
func lastTradingDay(now time.Time) time.Time {
	day := now.Truncate(24 * time.Hour).Add(-24 * time.Hour)
	switch day.Weekday() {
	case time.Sunday:
		day = day.Add(-2 * 24 * time.Hour)
	case time.Saturday:
		day = day.Add(-1 * 24 * time.Hour)
	}
	return day
}

// periodCutoff converts a period string to an absolute cutoff date.
func periodCutoff(period string) time.Time {
	return time.Now().UTC().Add(-periodToDuration(period))
}

func periodToDuration(period string) time.Duration {
	switch period {
	case "1m":
		return 30 * 24 * time.Hour
	case "3m":
		return 90 * 24 * time.Hour
	case "6m":
		return 180 * 24 * time.Hour
	case "1y":
		return 365 * 24 * time.Hour
	case "2y":
		return 730 * 24 * time.Hour
	default:
		return 365 * 24 * time.Hour
	}
}
