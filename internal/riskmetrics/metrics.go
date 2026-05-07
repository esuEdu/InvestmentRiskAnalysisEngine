package riskmetrics

import "time"

type PricePoint struct {
	Date  time.Time `json:"date"`
	Close float64   `json:"close"`
}

type Asset struct {
	Ticker string
	Prices []PricePoint
	Weight float64
}

type MetricsInput struct {
	Assets          []Asset
	BenchmarkPrices []PricePoint
	RiskFreeRate    float64
}

type MetricsResult struct {
	AnnualizedVolatility float64  `json:"annualized_volatility"`
	SharpeRatio          float64  `json:"sharpe_ratio"`
	Beta                 *float64 `json:"beta"`
	MaxDrawdown          float64  `json:"max_drawdown"`
	VaR95                float64  `json:"var_95"`
	ConcentrationScore   float64  `json:"concentration_score"`
}

type Calculator interface {
	Compute(input MetricsInput) (MetricsResult, error)
}
