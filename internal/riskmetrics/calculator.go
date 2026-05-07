package riskmetrics

import (
	"fmt"
	"math"
	"sort"
)

type DefaultCalculator struct{}

func NewCalculator() *DefaultCalculator {
	return &DefaultCalculator{}
}

func (c *DefaultCalculator) Compute(input MetricsInput) (MetricsResult, error) {
	if len(input.Assets) == 0 {
		return MetricsResult{}, fmt.Errorf("no assets provided")
	}

	// Build daily log returns per asset keyed by date string.
	assetReturns := make([]map[string]float64, len(input.Assets))
	for i, asset := range input.Assets {
		r, err := dailyLogReturns(asset.Prices)
		if err != nil {
			return MetricsResult{}, fmt.Errorf("asset %s: %w", asset.Ticker, err)
		}
		assetReturns[i] = r
	}

	// Find the intersection of all asset return dates.
	dates := returnDateIntersection(assetReturns)
	if len(dates) < 2 {
		return MetricsResult{}, fmt.Errorf("insufficient overlapping price data (need at least 2 common trading days)")
	}

	// Combine into weighted portfolio returns.
	portReturns := make([]float64, len(dates))
	for i, d := range dates {
		for j, asset := range input.Assets {
			portReturns[i] += asset.Weight * assetReturns[j][d]
		}
	}

	vol := annualizedVolatility(portReturns)
	sharpe := sharpeRatio(portReturns, vol, input.RiskFreeRate)
	mdd := maxDrawdown(input.Assets[0].Prices)
	var95 := historicalVaR(portReturns, 0.05)
	hhi := concentrationScore(input.Assets)

	var beta *float64
	if len(input.BenchmarkPrices) > 1 {
		benchReturns, err := dailyLogReturns(input.BenchmarkPrices)
		if err == nil {
			b := computeBeta(portReturns, dates, benchReturns)
			beta = &b
		}
	}

	return MetricsResult{
		AnnualizedVolatility: vol,
		SharpeRatio:          sharpe,
		Beta:                 beta,
		MaxDrawdown:          mdd,
		VaR95:                var95,
		ConcentrationScore:   hhi,
	}, nil
}

// dailyLogReturns computes ln(close_t / close_{t-1}) for each consecutive pair.
func dailyLogReturns(prices []PricePoint) (map[string]float64, error) {
	if len(prices) < 2 {
		return nil, fmt.Errorf("need at least 2 price points")
	}
	returns := make(map[string]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1].Close <= 0 {
			continue
		}
		returns[prices[i].Date.Format("2006-01-02")] = math.Log(prices[i].Close / prices[i-1].Close)
	}
	return returns, nil
}

// returnDateIntersection returns sorted date strings common to all return maps.
func returnDateIntersection(returns []map[string]float64) []string {
	if len(returns) == 0 {
		return nil
	}
	common := make(map[string]bool)
	for d := range returns[0] {
		common[d] = true
	}
	for _, r := range returns[1:] {
		for d := range common {
			if _, ok := r[d]; !ok {
				delete(common, d)
			}
		}
	}
	dates := make([]string, 0, len(common))
	for d := range common {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	return dates
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func stddev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	variance := 0.0
	for _, x := range xs {
		d := x - m
		variance += d * d
	}
	return math.Sqrt(variance / float64(len(xs)-1))
}

func annualizedVolatility(portReturns []float64) float64 {
	return stddev(portReturns) * math.Sqrt(252)
}

func sharpeRatio(portReturns []float64, vol, riskFreeRate float64) float64 {
	if vol == 0 {
		return 0
	}
	annualReturn := mean(portReturns) * 252
	return (annualReturn - riskFreeRate) / vol
}

// maxDrawdown returns the largest peak-to-trough decline as a positive value.
func maxDrawdown(prices []PricePoint) float64 {
	if len(prices) == 0 {
		return 0
	}
	peak := prices[0].Close
	maxDD := 0.0
	for _, p := range prices[1:] {
		if p.Close > peak {
			peak = p.Close
		}
		if peak > 0 {
			dd := (peak - p.Close) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD
}

// historicalVaR returns the q-th percentile of returns (e.g. q=0.05 for 95% VaR).
func historicalVaR(portReturns []float64, q float64) float64 {
	if len(portReturns) == 0 {
		return 0
	}
	sorted := make([]float64, len(portReturns))
	copy(sorted, portReturns)
	sort.Float64s(sorted)
	idx := int(math.Floor(q * float64(len(sorted))))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// concentrationScore computes the Herfindahl-Hirschman Index: Σ w_i².
func concentrationScore(assets []Asset) float64 {
	hhi := 0.0
	for _, a := range assets {
		hhi += a.Weight * a.Weight
	}
	return hhi
}

// computeBeta computes Cov(portfolio, benchmark) / Var(benchmark).
func computeBeta(portReturns []float64, dates []string, benchReturns map[string]float64) float64 {
	bench := make([]float64, 0, len(dates))
	port := make([]float64, 0, len(dates))
	for i, d := range dates {
		if b, ok := benchReturns[d]; ok {
			bench = append(bench, b)
			port = append(port, portReturns[i])
		}
	}
	if len(bench) < 2 {
		return 0
	}
	mPort := mean(port)
	mBench := mean(bench)
	cov := 0.0
	varB := 0.0
	for i := range bench {
		cov += (port[i] - mPort) * (bench[i] - mBench)
		varB += (bench[i] - mBench) * (bench[i] - mBench)
	}
	if varB == 0 {
		return 0
	}
	return cov / varB
}
