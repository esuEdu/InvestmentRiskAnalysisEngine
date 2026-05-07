package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/marketdata/domain"
	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	"golang.org/x/time/rate"
)

const defaultBaseURL = "https://www.alphavantage.co/query"

type AlphaVantageClient struct {
	httpClient *http.Client
	apiKey     string
	limiter    *rate.Limiter
	baseURL    string
}

func New(apiKey string) *AlphaVantageClient {
	return &AlphaVantageClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		// 5 requests per minute = 1 token every 12 seconds
		limiter: rate.NewLimiter(rate.Every(12*time.Second), 1),
		baseURL: defaultBaseURL,
	}
}

// avRawResponse captures both the happy-path time series and any error fields
// AlphaVantage may return instead of data.
type avRawResponse struct {
	// Error fields returned when quota is exceeded or key is invalid.
	Information  string `json:"Information"`
	Note         string `json:"Note"`
	ErrorMessage string `json:"Error Message"`

	// Actual price data.
	TimeSeries map[string]avBar `json:"Time Series (Daily)"`
}

type avBar struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

func (c *AlphaVantageClient) FetchDailyPrices(ctx context.Context, ticker string) ([]domain.PricePoint, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	// compact = last 100 trading days (~5 months), available on the free plan.
	// full    = up to 20 years, requires a premium AlphaVantage subscription.
	outputSize := "compact"

	url := fmt.Sprintf("%s?function=TIME_SERIES_DAILY&symbol=%s&outputsize=%s&apikey=%s",
		c.baseURL, ticker, outputSize, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("alphavantage returned HTTP %d for %s", resp.StatusCode, ticker)
	}

	var raw avRawResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode alphavantage response: %w", err)
	}

	// Surface explicit API error messages before the empty-map check.
	if raw.ErrorMessage != "" {
		return nil, fmt.Errorf("alphavantage error for %s: %s", ticker, raw.ErrorMessage)
	}
	if raw.Information != "" {
		logger.Log.Warnw("AlphaVantage quota/info message", "ticker", ticker, "message", raw.Information)
		return nil, fmt.Errorf("alphavantage quota exceeded or key not active for %s: %s", ticker, raw.Information)
	}
	if raw.Note != "" {
		logger.Log.Warnw("AlphaVantage rate limit note", "ticker", ticker, "note", raw.Note)
		return nil, fmt.Errorf("alphavantage rate limit for %s: %s", ticker, raw.Note)
	}

	if len(raw.TimeSeries) == 0 {
		return nil, fmt.Errorf("alphavantage returned no price data for %s", ticker)
	}

	prices := make([]domain.PricePoint, 0, len(raw.TimeSeries))
	for dateStr, bar := range raw.TimeSeries {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		prices = append(prices, domain.PricePoint{
			Date:   date,
			Open:   parseFloat(bar.Open),
			High:   parseFloat(bar.High),
			Low:    parseFloat(bar.Low),
			Close:  parseFloat(bar.Close),
			Volume: parseInt(bar.Volume),
		})
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.Before(prices[j].Date)
	})

	logger.Log.Infow("AlphaVantage fetch complete", "ticker", ticker, "points", len(prices))
	return prices, nil
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
