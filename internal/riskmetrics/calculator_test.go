package riskmetrics_test

import (
	"math"
	"testing"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/riskmetrics"
)

func makePrice(daysAgo int, close float64) riskmetrics.PricePoint {
	return riskmetrics.PricePoint{
		Date:  time.Now().AddDate(0, 0, -daysAgo),
		Close: close,
	}
}

// linearPrices generates n price points with a constant daily step.
func linearPrices(n int, start, step float64) []riskmetrics.PricePoint {
	prices := make([]riskmetrics.PricePoint, n)
	for i := 0; i < n; i++ {
		prices[i] = riskmetrics.PricePoint{
			Date:  time.Now().AddDate(0, 0, -(n - 1 - i)),
			Close: start + float64(i)*step,
		}
	}
	return prices
}

func TestCompute_SingleAsset_FlatPrices(t *testing.T) {
	// Flat prices → zero returns → zero volatility and zero Sharpe.
	prices := linearPrices(10, 100, 0)
	calc := riskmetrics.NewCalculator()

	result, err := calc.Compute(riskmetrics.MetricsInput{
		Assets:       []riskmetrics.Asset{{Ticker: "FLAT", Prices: prices, Weight: 1.0}},
		RiskFreeRate: 0.05,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AnnualizedVolatility != 0 {
		t.Errorf("want volatility=0 for flat prices, got %f", result.AnnualizedVolatility)
	}
	if result.MaxDrawdown != 0 {
		t.Errorf("want max drawdown=0 for flat prices, got %f", result.MaxDrawdown)
	}
}

func TestCompute_SingleAsset_KnownDrawdown(t *testing.T) {
	// Peak at 200, trough at 100 → max drawdown = 0.5.
	prices := []riskmetrics.PricePoint{
		makePrice(4, 100),
		makePrice(3, 200),
		makePrice(2, 150),
		makePrice(1, 100),
		makePrice(0, 120),
	}
	calc := riskmetrics.NewCalculator()

	result, err := calc.Compute(riskmetrics.MetricsInput{
		Assets:       []riskmetrics.Asset{{Ticker: "X", Prices: prices, Weight: 1.0}},
		RiskFreeRate: 0.05,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const want = 0.5
	if math.Abs(result.MaxDrawdown-want) > 1e-9 {
		t.Errorf("want max drawdown=%f, got %f", want, result.MaxDrawdown)
	}
}

func TestCompute_HHI_TwoEqualWeights(t *testing.T) {
	// Two assets each with weight 0.5 → HHI = 0.5² + 0.5² = 0.5
	prices := linearPrices(5, 100, 1)
	calc := riskmetrics.NewCalculator()

	result, err := calc.Compute(riskmetrics.MetricsInput{
		Assets: []riskmetrics.Asset{
			{Ticker: "A", Prices: prices, Weight: 0.5},
			{Ticker: "B", Prices: prices, Weight: 0.5},
		},
		RiskFreeRate: 0.05,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const want = 0.5
	if math.Abs(result.ConcentrationScore-want) > 1e-9 {
		t.Errorf("want HHI=%f, got %f", want, result.ConcentrationScore)
	}
}

func TestCompute_NoBenchmark_BetaIsNil(t *testing.T) {
	prices := linearPrices(5, 100, 2)
	calc := riskmetrics.NewCalculator()

	result, err := calc.Compute(riskmetrics.MetricsInput{
		Assets:       []riskmetrics.Asset{{Ticker: "X", Prices: prices, Weight: 1.0}},
		RiskFreeRate: 0.05,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Beta != nil {
		t.Errorf("want Beta=nil when no benchmark provided, got %v", *result.Beta)
	}
}

func TestCompute_WithBenchmark_BetaNotNil(t *testing.T) {
	prices := linearPrices(10, 100, 1)
	calc := riskmetrics.NewCalculator()

	result, err := calc.Compute(riskmetrics.MetricsInput{
		Assets:          []riskmetrics.Asset{{Ticker: "X", Prices: prices, Weight: 1.0}},
		BenchmarkPrices: prices,
		RiskFreeRate:    0.05,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Beta == nil {
		t.Error("want Beta to be set when benchmark provided")
	}
}

func TestCompute_NoAssets_Error(t *testing.T) {
	calc := riskmetrics.NewCalculator()
	_, err := calc.Compute(riskmetrics.MetricsInput{RiskFreeRate: 0.05})
	if err == nil {
		t.Fatal("expected error for empty assets, got nil")
	}
}
