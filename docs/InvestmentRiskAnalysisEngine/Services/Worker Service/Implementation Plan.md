# Implementation Plan — Worker Service (Phase 3)

> Approved implementation plan for the Worker Service, Market Data Service, and Risk Metrics package.

---

## What We Are Building

| Service / Package | Type | Purpose |
|---|---|---|
| `cmd/marketdata` | **Service** (:8081) | Cache-first price fetcher — AlphaVantage → PostgreSQL |
| `cmd/worker` | **Service** (no HTTP) | RabbitMQ consumer — orchestrates risk calculation per job |
| `internal/riskmetrics` | **Internal package** | Pure math — volatility, Sharpe, beta, VaR, drawdown, HHI |
| API extensions | **Existing service** | Multi-ticker POST body + `GET /analyses/:id/results` endpoint |

---

## Execution Order

Dependencies flow downward — each step must complete before the next begins.

```
Step 1 → DB Migrations (3 new tables)
Step 2 → SQL Queries + make generate
Step 3 → Config extensions
Step 4 → Analysis domain model changes (Asset, AnalysisResult, Assets field)
Step 5 → Analysis repository (transaction support + new methods)
Step 6 → Analysis usecase + handler + router (POST assets, GET results)
Step 7 → Middleware refactor (ZapLogger → pkg/middleware)
Step 8 → internal/marketdata (domain → repository → provider → usecase → handler)
Step 9 → cmd/marketdata/main.go
Step 10 → internal/riskmetrics (pure math, with unit tests)
Step 11 → internal/worker (handler + HTTP market data client)
Step 12 → cmd/worker/main.go
Step 13 → Docker Compose + Dockerfiles
```

---

## Step 1 — DB Migrations

Create 6 files in `internal/db/migrations/`:

### `000002_create_analysis_request_assets.up.sql`
```sql
CREATE TABLE IF NOT EXISTS analysis_request_assets (
    analysis_request_id UUID NOT NULL REFERENCES analysis_requests(id) ON DELETE CASCADE,
    ticker              TEXT NOT NULL,
    weight              NUMERIC(6,4) NOT NULL,
    PRIMARY KEY (analysis_request_id, ticker)
);
```

### `000003_create_analysis_results.up.sql`
```sql
CREATE TABLE IF NOT EXISTS analysis_results (
    analysis_request_id   UUID PRIMARY KEY REFERENCES analysis_requests(id) ON DELETE CASCADE,
    annualized_volatility NUMERIC NOT NULL,
    sharpe_ratio          NUMERIC NOT NULL,
    beta                  NUMERIC,
    max_drawdown          NUMERIC NOT NULL,
    var_95                NUMERIC NOT NULL,
    concentration_score   NUMERIC NOT NULL,
    raw_metrics_json      JSONB,
    created_at            TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### `000004_create_historical_prices.up.sql`
```sql
CREATE TABLE IF NOT EXISTS historical_prices (
    ticker     TEXT NOT NULL,
    price_date DATE NOT NULL,
    open       NUMERIC, high NUMERIC, low NUMERIC,
    close      NUMERIC NOT NULL,
    volume     BIGINT,
    PRIMARY KEY (ticker, price_date)
);
CREATE INDEX idx_historical_prices_ticker_date ON historical_prices(ticker, price_date DESC);
```

Each gets a matching `.down.sql`. Run `make migrate-up` after all 6 files are created.

---

## Step 2 — SQL Queries + sqlc

Create 3 new query files in `internal/db/queries/`:

### `analysis_request_assets.sql`
```sql
-- name: InsertAnalysisRequestAsset :exec
INSERT INTO analysis_request_assets (analysis_request_id, ticker, weight)
VALUES ($1, $2, $3);

-- name: GetAnalysisRequestAssets :many
SELECT ticker, weight FROM analysis_request_assets
WHERE analysis_request_id = $1 ORDER BY ticker;
```

### `analysis_results.sql`
```sql
-- name: InsertAnalysisResult :exec
INSERT INTO analysis_results (
    analysis_request_id, annualized_volatility, sharpe_ratio,
    beta, max_drawdown, var_95, concentration_score, raw_metrics_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetAnalysisResult :one
SELECT * FROM analysis_results WHERE analysis_request_id = $1;
```

### `historical_prices.sql`
```sql
-- name: UpsertHistoricalPrice :exec
INSERT INTO historical_prices (ticker, price_date, open, high, low, close, volume)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (ticker, price_date) DO UPDATE
SET open=EXCLUDED.open, high=EXCLUDED.high, low=EXCLUDED.low,
    close=EXCLUDED.close, volume=EXCLUDED.volume;

-- name: GetHistoricalPrices :many
SELECT ticker, price_date, open, high, low, close, volume
FROM historical_prices WHERE ticker = $1 ORDER BY price_date ASC;

-- name: GetLatestPriceDate :one
SELECT price_date FROM historical_prices
WHERE ticker = $1 ORDER BY price_date DESC LIMIT 1;
```

Run `make generate` → new generated files in `internal/db/generated/`.

---

## Step 3 — Config

**Edit `internal/config/config.go`** — add two fields:
```go
MarketDataServiceURL string  // env: MARKET_DATA_SERVICE_URL
MarketDataAPIKey     string  // env: MARKET_DATA_API_KEY
```

**Edit `.env`**:
```
MARKET_DATA_SERVICE_URL=http://localhost:8081
MARKET_DATA_API_KEY=your_alphavantage_key
```

---

## Step 4 — Extend Analysis Domain Model

**Edit `internal/analysis/domain/model.go`**

Add `Asset` struct (JSON tags required — serialized into RabbitMQ message):
```go
type Asset struct {
    Ticker string  `json:"ticker"`
    Weight float64 `json:"weight"`
}
```

Add `Assets []Asset` to `AnalysisRequest`. Consumer automatically deserializes it — no consumer changes needed.

Add `AnalysisResult` struct (in domain to avoid import cycle with riskmetrics):
```go
type AnalysisResult struct {
    AnnualizedVolatility float64
    SharpeRatio          float64
    Beta                 *float64
    MaxDrawdown          float64
    VaR95                float64
    ConcentrationScore   float64
    RawMetricsJSON       []byte
}
```

**Edit `internal/analysis/domain/repository.go`** — add to `Repository` interface:
```go
GetAssets(ctx context.Context, id uuid.UUID) ([]Asset, error)
GetAnalysisResult(ctx context.Context, id uuid.UUID) (AnalysisResult, error)
CreateAnalysisResult(ctx context.Context, id uuid.UUID, result AnalysisResult) error
```

---

## Step 5 — Extend Analysis Repository

**Edit `internal/analysis/repository/repository.go`**

Add `pool *pgxpool.Pool` to `Repo` struct — required for DB transactions:
```go
type Repo struct {
    queries *sqlc.Queries
    pool    *pgxpool.Pool
}
func New(q *sqlc.Queries, pool *pgxpool.Pool) *Repo
```

Update `Create()` to use a transaction (atomically writes request + assets):
```go
tx, _ := r.pool.Begin(ctx)
defer tx.Rollback(ctx)
qtx := r.queries.WithTx(tx)
// insert analysis_request
// for each asset: insert asset row
tx.Commit(ctx)
```

Add `GetAssets()`, `GetAnalysisResult()`, `CreateAnalysisResult()` using generated sqlc methods.

**Edit `internal/analysis/repository/mapper.go`** — add:
- `pgNumeric(f float64) pgtype.Numeric`
- `numericToFloat64(n pgtype.Numeric) float64`

**Edit `cmd/api/main.go`** — update: `repository.New(queries, pool)`

---

## Step 6 — Analysis UseCase + Handler + Router

**Edit `internal/analysis/usecase/usecase.go`**
- Update `ExecuteCreate` to accept `assets []domain.Asset`
- Add `ExecuteGetResult(ctx, id) (domain.AnalysisResult, error)`

**Edit `internal/analysis/delivery/http/handler.go`**
- Replace `PortfolioID` with `Assets []AssetRequest` in `CreateRequest`
- Validate weight sum: `math.Abs(total - 1.0) > 1e-4 → 400 Bad Request`
- Add `GetResult(c *gin.Context)` handler → 200 with metrics or 404

**Edit `internal/server/routers.go`**:
```go
analysis.GET("/:id/results", s.analysisHandler.GetResult)
```

---

## Step 7 — Middleware Refactor

**Create `pkg/middleware/zap.go`** — move `ZapLogger()` from `internal/server/middleware.go` here.

**Edit `internal/server/middleware.go`** — delegate to `pkg/middleware.ZapLogger()`.

This allows `cmd/marketdata/main.go` to use the same middleware without importing `internal/server`.

---

## Step 8 — internal/marketdata (Clean Architecture)

Same structure as `internal/analysis/`:

```
internal/marketdata/
  domain/
    model.go        ← PricePoint{Date, Open, High, Low, Close, Volume}
    repository.go   ← Repository interface (GetPrices, GetLatestPriceDate, UpsertPrices)
    provider.go     ← Provider interface (FetchDailyPrices)
  repository/
    repository.go   ← sqlc implementation
    mapper.go       ← pgtype.Date→time.Time, pgtype.Numeric→float64
  provider/
    alphaVantage.go ← HTTP adapter + rate limiter
  usecase/
    usecase.go      ← GetPrices: freshness check → fetch if stale → filter by period → return
  delivery/http/
    handler.go      ← GET /api/v1/prices?ticker=X&period=Y
```

### AlphaVantage Provider (key details)
- Rate limiter: `rate.NewLimiter(rate.Every(12*time.Second), 1)` (`golang.org/x/time/rate`)
- `limiter.Wait(ctx)` before every HTTP call
- URL: `TIME_SERIES_DAILY`, `outputsize=full`
- Parse `"Time Series (Daily)"` JSON key with local private struct

### Usecase Freshness Logic
- Fresh if `latestPriceDate ≥ yesterday` (yesterday = last Friday if today is Monday)
- Period filter: `"1y"=365d`, `"6m"=180d`, `"3m"=90d`, `"1m"=30d` → cutoff = now − duration

---

## Step 9 — cmd/marketdata/main.go

Gin HTTP server on `:8081`. Wiring:
```
config → pool → sqlc.New → marketdata/repository.New → alphaVantage.New(cfg.MarketDataAPIKey)
→ usecase.New(repo, provider) → handler.New(uc) → gin → r.Run(":8081")
```
`signal.NotifyContext(SIGINT, SIGTERM)` for graceful shutdown.

---

## Step 10 — internal/riskmetrics

Pure math. No DB, no HTTP. Only stdlib imports.

**Types** (`metrics.go`):
```go
type PricePoint struct { Date time.Time `json:"date"`; Close float64 `json:"close"` }
type Asset struct { Ticker string; Prices []PricePoint; Weight float64 }
type MetricsInput struct { Assets []Asset; BenchmarkPrices []PricePoint; RiskFreeRate float64 }
type MetricsResult struct {
    AnnualizedVolatility, SharpeRatio, MaxDrawdown, VaR95, ConcentrationScore float64
    Beta *float64
}
type Calculator interface { Compute(MetricsInput) (MetricsResult, error) }
```

**`DefaultCalculator.Compute` algorithm** (`calculator.go`):
1. Daily log returns per asset: `math.Log(close_t / close_{t-1})`
2. Align returns by date — intersection of all asset return dates
3. Portfolio return/day: `Σ w_i × r_i_t`
4. Volatility: `stddev(portfolio_returns) × √252`
5. Sharpe: `(mean(returns)×252 − 0.05) / volatility`
6. Beta: `Cov(portfolio, benchmark) / Var(benchmark)` — `nil` if no benchmark
7. Max drawdown: track running peak, `min((p−peak)/peak)` as positive
8. VaR95: sort returns asc → index `floor(0.05 × n)`
9. HHI: `Σ w_i²`

Unit tests with synthetic price series in `calculator_test.go`.

---

## Step 11 — internal/worker

**`worker.go`** — `Handler` with local interfaces (testability):
```go
type MarketDataClient interface {
    GetPrices(ctx context.Context, ticker, period string) ([]riskmetrics.PricePoint, error)
}
type AnalysisRepository interface {
    UpdateStatus(ctx, id, status) error
    GetAssets(ctx, id) ([]domain.Asset, error)
    CreateAnalysisResult(ctx, id, result) error
}
```

`Handle(ctx, req)` flow:
1. `UpdateStatus → processing`
2. `GetAssets(req.ID)` from DB
3. Per asset: `mdClient.GetPrices(ticker, period)`
4. If benchmark: `mdClient.GetPrices(benchmark, period)`
5. `calculator.Compute(input)` with `RiskFreeRate: 0.05`
6. `json.Marshal(result)` → `CreateAnalysisResult`
7. `UpdateStatus → completed` (or `failed` on any error)

**`marketdata_client.go`** — `HTTPMarketDataClient`:
- `GET {baseURL}/api/v1/prices?ticker=X&period=Y`
- Decodes `{ "data": [...] }` envelope (standard `response.Body`)
- Uses `http.NewRequestWithContext` for proper cancellation

---

## Step 12 — cmd/worker/main.go

Wiring:
```
config → pool → mq → sqlc.New → analysis/repository.New(queries, pool)
→ worker.NewHTTPMarketDataClient(cfg.MarketDataServiceURL)
→ riskmetrics.NewCalculator()
→ worker.New(repo, mdClient, calculator)
→ consumer.NewAnalysisConsumer(consumer.NewConsumer(mq.Conn))
→ analysisConsumer.Start(ctx, workerHandler.Handle)  ← blocks until SIGTERM
```

---

## Step 13 — Docker Compose + Dockerfiles

**Edit `docker-compose.yml`** — add `marketdata` (port 8081, depends on postgres) and `worker` (no port, depends on postgres + rabbitMQ + marketdata, `MARKET_DATA_SERVICE_URL=http://marketdata:8081`).

**Create `deployments/docker/Dockerfile.marketdata`** and **`Dockerfile.worker`** — multi-stage builds.

---

## Files Summary

### New files (24)
- `internal/db/migrations/000002_*.{up,down}.sql` (2)
- `internal/db/migrations/000003_*.{up,down}.sql` (2)
- `internal/db/migrations/000004_*.{up,down}.sql` (2)
- `internal/db/queries/analysis_request_assets.sql`
- `internal/db/queries/analysis_results.sql`
- `internal/db/queries/historical_prices.sql`
- `pkg/middleware/zap.go`
- `internal/marketdata/domain/model.go`
- `internal/marketdata/domain/repository.go`
- `internal/marketdata/domain/provider.go`
- `internal/marketdata/repository/repository.go`
- `internal/marketdata/repository/mapper.go`
- `internal/marketdata/provider/alphaVantage.go`
- `internal/marketdata/usecase/usecase.go`
- `internal/marketdata/delivery/http/handler.go`
- `cmd/marketdata/main.go`
- `internal/riskmetrics/metrics.go`
- `internal/riskmetrics/calculator.go`
- `internal/riskmetrics/calculator_test.go`
- `internal/worker/worker.go`
- `internal/worker/marketdata_client.go`
- `cmd/worker/main.go`
- `deployments/docker/Dockerfile.marketdata`
- `deployments/docker/Dockerfile.worker`

### Existing files to edit (12)
- `internal/analysis/domain/model.go`
- `internal/analysis/domain/repository.go`
- `internal/analysis/repository/repository.go`
- `internal/analysis/repository/mapper.go`
- `internal/analysis/usecase/usecase.go`
- `internal/analysis/delivery/http/handler.go`
- `internal/server/routers.go`
- `internal/server/middleware.go`
- `internal/config/config.go`
- `cmd/api/main.go`
- `docker-compose.yml`
- `.env`

---

## Verification Checklist

- [ ] `make generate` — succeeds
- [ ] `make migrate-up` — all 4 migrations run
- [ ] `go build ./...` — full monorepo compiles
- [ ] `go test ./internal/riskmetrics/...` — calculator tests pass
- [ ] `go test ./internal/analysis/...` — existing tests pass
- [ ] `curl "http://localhost:8081/api/v1/prices?ticker=AAPL&period=1y"` — returns price array
- [ ] E2E: POST analyses with assets → 202 → worker logs processing → GET results → metrics returned

---

## Related Notes

- [[Architecture]]
- [[Worker Service]]
- [[Database Schema]]
- [[API Reference]]
- [[Project Plan]]
