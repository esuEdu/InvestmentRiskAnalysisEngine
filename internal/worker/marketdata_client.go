package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/riskmetrics"
)

type HTTPMarketDataClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPMarketDataClient(baseURL string) *HTTPMarketDataClient {
	return &HTTPMarketDataClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPMarketDataClient) GetPrices(ctx context.Context, ticker, period string) ([]riskmetrics.PricePoint, error) {
	url := fmt.Sprintf("%s/api/v1/prices?ticker=%s&period=%s", c.baseURL, ticker, period)

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
		return nil, fmt.Errorf("market data service returned status %d for ticker %s", resp.StatusCode, ticker)
	}

	var envelope struct {
		Data []riskmetrics.PricePoint `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode market data response: %w", err)
	}

	return envelope.Data, nil
}
