# Worker Service

> Implementation reference for `cmd/worker` — Phase 3 of the Investment Risk Analysis Engine (complete).

---

## Services vs Functions — Clarification

### What is a Service?

A **service** is its own binary (`cmd/<name>/main.go`) running as a **separate process**. It has its own lifecycle, can be deployed and scaled independently, and communicates with other services over the network (HTTP or RabbitMQ).

### What is an Internal Package (Functions)?

An **internal package** is a Go library inside `internal/`. It is **not** a process — it is compiled *into* a service binary. It has no network presence. It is just reusable logic that one or more services import.

---

## Services in this System

| Service | Binary | Port | Responsibility |
|---|---|---|---|
| **API** | `cmd/api` | `:8080` | HTTP entry point — accepts requests, queues jobs |
| **Worker** | `cmd/worker` | — | Consumes jobs, orchestrates risk calculation |
| **Market Data** | `cmd/marketdata` | `:8081` | Fetches + caches prices; future: chart endpoints |
| **Portfolio** | `cmd/portfolio` *(future)* | `:8082` | Portfolios, asset weights, user profiles |

All services live in the **same monorepo** (`github.com/esuEdu/investment-risk-engine`).

---

## Internal Packages (Functions, NOT Services)

| Package | Used by | Responsibility |
|---|---|---|
| `internal/riskmetrics` | Worker | Pure math: volatility, Sharpe, beta, VaR, drawdown, HHI |
| `internal/analysis` | API + Worker | Domain model, repository, use cases |
| `internal/messaging` | API + Worker | RabbitMQ connection, publisher, consumer |
| `internal/config` | All services | Viper config loader |
| `internal/db` | All services | pgx pool + sqlc-generated queries |
| `pkg/logger` | All services | Uber Zap singleton |

---

## System Architecture

```
Client
  │
  ▼
API Service (:8080)
  │  POST /api/v1/analyses
  │  → saves analysis_requests + analysis_request_assets
  │  → publishes job to RabbitMQ
  ▼
RabbitMQ — "risk-analysis-jobs"
  │
  ▼
Worker Service (no HTTP)
  │  for each asset ticker:
  │    GET http://marketdata:8081/prices?ticker=X&period=Y
  │  combine weighted returns → riskmetrics.Compute()
  │  INSERT analysis_results
  │  UPDATE analysis_requests.status → completed/failed
  │
  ▼
Market Data Service (:8081)
  checks historical_prices in DB (cache)
  if stale → AlphaVantage API (rate-limited: 5 req/min, 25 req/day)
  returns []PricePoint to Worker

GET /api/v1/analyses/:id/results
  → API reads analysis_results and returns to client
```

---

## Decisions (Closed)

| Question | Decision |
|---|---|
| Market data provider | **AlphaVantage Free Plan** — 25 req/day, 5 req/min |
| Risk-free rate | **Hardcode `0.05`** (5%) |
| Single vs multi-ticker | **Multi-ticker** — portfolio of N assets with weights |
| Results via API | **Yes** — `GET /api/v1/analyses/:id/results` |
| Monorepo | **Yes** — all services under `cmd/` in the same repo |

---

## Multi-Ticker Analysis

### Request Model Change

`analysis_requests` gets a companion table `analysis_request_assets`:

```
POST /api/v1/analyses body:
{
  "benchmark": "SPY",
  "period": "1y",
  "assets": [
    { "ticker": "AAPL", "weight": 0.40 },
    { "ticker": "MSFT", "weight": 0.35 },
    { "ticker": "GOOGL", "weight": 0.25 }
  ]
}
```

Weights must sum to `1.0` — validated by the API handler.

### Portfolio Return Calculation (inside Worker)

For each trading day `t`:

```
portfolio_return_t = Σ (w_i × daily_log_return_i_t)
```

All six risk metrics are then computed on this combined `portfolio_return` series. This is what `internal/riskmetrics` receives.

---

## AlphaVantage Rate Limiting

**Free Plan constraints:**
- 5 requests per minute
- 25 requests per day

**Impact on multi-ticker:** A 5-asset portfolio = 5 API calls if none are cached. A 25-asset portfolio would exhaust the daily quota in one job.

**Strategy — cache-first in Market Data Service:**

```
GET /prices?ticker=AAPL&period=1y
  │
  ├─ Query historical_prices WHERE ticker='AAPL'
  │    AND price_date >= NOW() - INTERVAL '1y'
  │
  ├─ If rows exist AND last row date = yesterday (market closed today)
  │    → return cached rows immediately (no API call)
  │
  └─ If missing or stale
       → fetch from AlphaVantage
       → upsert into historical_prices
       → return rows
       (costs 1 API request)
```

**Rate limiter in Market Data Service:**

```go
// golang.org/x/time/rate — token bucket
// 5 req/min = 1 token every 12 seconds
limiter := rate.NewLimiter(rate.Every(12*time.Second), 1)

func (p *AlphaVantageProvider) Fetch(ctx context.Context, ticker, period string) ([]PricePoint, error) {
    limiter.Wait(ctx)
    // ... HTTP call to AlphaVantage
}
```

**Daily quota guard:** If the daily budget is exhausted, return an error — the Worker marks the job `failed` and it is requeued for the next day.

---

## Worker — Data Flow per Job

```
AnalysisRequest consumed from RabbitMQ
  │
  ├─ 1. repo.UpdateStatus(id, "processing")
  │
  ├─ 2. repo.GetAssets(id) → []Asset{ticker, weight}
  │
  ├─ 3. For each asset:
  │        marketClient.GetPrices(ctx, ticker, period)
  │        → []PricePoint
  │        (Market Data Service handles cache + rate limit)
  │
  ├─ 4. If benchmark set:
  │        marketClient.GetPrices(ctx, benchmark, period)
  │        → []PricePoint (for Beta)
  │
  ├─ 5. riskmetrics.Compute(MetricsInput{assets, benchmark, 0.05})
  │        → MetricsResult
  │
  ├─ 6. repo.CreateAnalysisResult(id, MetricsResult)
  │
  └─ 7. repo.UpdateStatus(id, "completed")
         (or "failed" on any error above)
```

---

## Package Layout

```
cmd/
  api/main.go              ← API Service
  worker/main.go           ← Worker Service
  marketdata/main.go       ← Market Data Service

internal/
  worker/
    worker.go              ← Handler: orchestrates steps 1–7 per job
    marketdata_client.go   ← HTTPMarketDataClient (calls marketdata over HTTP)

  marketdata/
    domain/
      model.go             ← PricePoint, OHLCV structs
      provider.go          ← Provider interface
      repository.go        ← Repository interface
    usecase/
      usecase.go           ← GetPrices — cache check + fetch if stale
    repository/
      mapper.go            ← DB row → domain type
      repository.go        ← reads/writes historical_prices
    delivery/http/
      handler.go           ← GET /prices?ticker=X&period=Y
    provider/
      alphaVantage.go      ← AlphaVantage adapter + rate limiter (5 req/min)

  riskmetrics/
    metrics.go             ← PricePoint, Asset, MetricsInput, MetricsResult, Calculator interface
    calculator.go          ← DefaultCalculator: all metrics in one pass
    calculator_test.go

  analysis/
    domain/
      model.go             ← Asset struct added, AnalysisRequest updated
    repository/
      repository.go        ← GetAssets, CreateAnalysisResult added
```

---

## Interfaces

### Market Data Client (HTTP adapter — used by Worker)

```go
// internal/marketdata/domain/model.go
type PricePoint struct {
    Date  time.Time
    Close float64
}

// internal/worker/worker.go (interface consumed by Worker)
type MarketDataClient interface {
    GetPrices(ctx context.Context, ticker, period string) ([]PricePoint, error)
}
```

### Risk Calculator (internal package — compiled into Worker)

```go
// internal/riskmetrics/metrics.go
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
    BenchmarkPrices []PricePoint // nil → Beta = nil
    RiskFreeRate    float64      // hardcoded 0.05
}

type MetricsResult struct {
    AnnualizedVolatility float64  `json:"annualized_volatility"`
    SharpeRatio          float64  `json:"sharpe_ratio"`
    Beta                 *float64 `json:"beta"` // nil if no benchmark
    MaxDrawdown          float64  `json:"max_drawdown"`
    VaR95                float64  `json:"var_95"`
    ConcentrationScore   float64  `json:"concentration_score"` // HHI = Σ w_i²
}

type Calculator interface {
    Compute(input MetricsInput) (MetricsResult, error)
}
```

### Worker Handler

```go
// internal/worker/worker.go
type Handler struct {
    repo       AnalysisRepository
    mdClient   MarketDataClient
    calculator riskmetrics.Calculator
}

func (h *Handler) Handle(ctx context.Context, req *domain.AnalysisRequest) error
```

---

## Entrypoint: `cmd/worker/main.go`

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    cfg := config.Load()
    logger.Initialize(cfg.AppEnv)

    pool, _  := db.NewPostgres(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
    mq, _    := messaging.NewRabbitMQ(ctx, cfg.MQHost, cfg.MQPort, cfg.MQUser, cfg.MQPassword)

    queries          := sqlc.New(pool)
    repo             := repository.New(queries, pool)
    mdClient         := worker.NewHTTPMarketDataClient(cfg.MarketDataServiceURL)
    calculator        := riskmetrics.NewCalculator()
    workerHandler    := worker.New(repo, mdClient, calculator)
    analysisConsumer := consumer.NewAnalysisConsumer(consumer.NewConsumer(mq.Conn))

    logger.Log.Infow("Worker Service started — listening on risk-analysis-jobs")
    analysisConsumer.Start(ctx, func(req *domain.AnalysisRequest) error {
        return workerHandler.Handle(ctx, req)
    })
}
```

---

## Database Changes

### New table: `analysis_request_assets`

```sql
CREATE TABLE analysis_request_assets (
    analysis_request_id UUID NOT NULL REFERENCES analysis_requests(id),
    ticker              TEXT NOT NULL,
    weight              NUMERIC(6, 4) NOT NULL CHECK (weight > 0 AND weight <= 1),
    PRIMARY KEY (analysis_request_id, ticker)
);
```

### New table: `analysis_results`

```sql
CREATE TABLE analysis_results (
    analysis_request_id   UUID PRIMARY KEY REFERENCES analysis_requests(id),
    annualized_volatility NUMERIC(10, 6),
    sharpe_ratio          NUMERIC(10, 6),
    beta                  NUMERIC(10, 6),
    max_drawdown          NUMERIC(10, 6),
    var_95                NUMERIC(10, 6),
    concentration_score   NUMERIC(10, 6),
    raw_metrics_json      JSONB,
    created_at            TIMESTAMPTZ DEFAULT NOW()
);
```

### New table: `historical_prices` (owned by Market Data Service)

```sql
CREATE TABLE historical_prices (
    ticker     TEXT    NOT NULL,
    price_date DATE    NOT NULL,
    open       NUMERIC(12, 4),
    high       NUMERIC(12, 4),
    low        NUMERIC(12, 4),
    close      NUMERIC(12, 4) NOT NULL,
    volume     BIGINT,
    PRIMARY KEY (ticker, price_date)
);

CREATE INDEX idx_historical_prices_ticker_date ON historical_prices(ticker, price_date DESC);
```

---

## Config Changes

```go
// internal/config/config.go additions
MarketDataServiceURL string  // env: MARKET_DATA_SERVICE_URL → http://marketdata:8081
MarketDataAPIKey     string  // env: MARKET_DATA_API_KEY → AlphaVantage key
```

---

## Error Handling

| Scenario | Action |
|---|---|
| Market Data 5xx / timeout | Return error → RabbitMQ requeues (up to 3x via DLQ) |
| Market Data 404 (bad ticker) | UpdateStatus `failed`, ACK — do not requeue |
| AlphaVantage daily quota hit | Return error → requeue for later |
| Risk calculation error | UpdateStatus `failed`, ACK |
| DB result insert fails | Return error → requeue |
| DB UpdateStatus fails | Log error, ACK — do not loop |

---

## Implementation Order (completed)

1. ✅ **DB migrations** — `analysis_request_assets`, `analysis_results`, `historical_prices` + `make generate`
2. ✅ **`internal/marketdata`** — AlphaVantage adapter with rate limiter, use case, repository, HTTP handler
3. ✅ **`cmd/marketdata/main.go`** — Market Data Service entrypoint
4. ✅ **`internal/riskmetrics`** — all calculators with unit tests
5. ✅ **`internal/worker`** — `Handle` orchestration with interfaces for testability
6. ✅ **`cmd/worker/main.go`** — Worker Service entrypoint with graceful shutdown
7. ✅ **API update** — multi-asset POST body + `GET /analyses/:id/results`
8. ✅ **Config** — `MarketDataServiceURL`, `MarketDataAPIKey`
9. ✅ **Docker Compose** — `worker` and `marketdata` services

---

## Related Notes

- [[Architecture]]
- [[Risk Metrics]]
- [[Database Schema]]
- [[API Reference]]
- [[Project Plan]]
