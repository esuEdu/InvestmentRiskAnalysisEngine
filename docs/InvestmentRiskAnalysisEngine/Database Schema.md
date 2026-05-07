# Database Schema

Generated with **sqlc** — queries live in `internal/db/generated/`.

---

## analysis_requests

Stores each incoming analysis job.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `benchmark` | TEXT | Optional ticker symbol used as benchmark (e.g. `SPY`) |
| `period` | TEXT | Time window (e.g. `1y`, `6m`, `3m`) |
| `status` | TEXT | `pending` / `processing` / `completed` / `failed` |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

---

## analysis_request_assets *(Phase 3 — new)*

Assets (tickers + weights) belonging to a single analysis request. Weights must sum to `1.0` — validated at the API layer.

| Column | Type | Notes |
|---|---|---|
| `analysis_request_id` | UUID | FK → `analysis_requests.id` |
| `ticker` | TEXT | Asset symbol (e.g. `AAPL`) |
| `weight` | NUMERIC(6,4) | Allocation weight, 0 < weight ≤ 1 |

**Primary key:** `(analysis_request_id, ticker)`

```sql
CREATE TABLE analysis_request_assets (
    analysis_request_id UUID NOT NULL REFERENCES analysis_requests(id),
    ticker              TEXT NOT NULL,
    weight              NUMERIC(6, 4) NOT NULL CHECK (weight > 0 AND weight <= 1),
    PRIMARY KEY (analysis_request_id, ticker)
);
```

---

## analysis_results *(Phase 3 — new)*

Computed risk metrics for a completed analysis. One row per `analysis_request_id`.

| Column | Type | Notes |
|---|---|---|
| `analysis_request_id` | UUID | PK + FK → `analysis_requests.id` |
| `annualized_volatility` | NUMERIC(10,6) | σ × √252 |
| `sharpe_ratio` | NUMERIC(10,6) | (return − 0.05) / σ |
| `beta` | NUMERIC(10,6) | vs benchmark; NULL if no benchmark given |
| `max_drawdown` | NUMERIC(10,6) | Largest peak-to-trough decline (negative value) |
| `var_95` | NUMERIC(10,6) | 5th-percentile daily return |
| `concentration_score` | NUMERIC(10,6) | Herfindahl index Σ w_i² |
| `raw_metrics_json` | JSONB | Full payload for future fields without migration |
| `created_at` | TIMESTAMPTZ | |

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

---

## historical_prices *(Phase 3 — new, owned by Market Data Service)*

Daily OHLCV price data cached from AlphaVantage. Other services must **not** query this table directly — go through the Market Data Service HTTP API.

| Column | Type | Notes |
|---|---|---|
| `ticker` | TEXT | Asset symbol |
| `price_date` | DATE | Trading day |
| `open` | NUMERIC(12,4) | |
| `high` | NUMERIC(12,4) | |
| `low` | NUMERIC(12,4) | |
| `close` | NUMERIC(12,4) | Required |
| `volume` | BIGINT | |

**Primary key:** `(ticker, price_date)`

```sql
CREATE TABLE historical_prices (
    ticker     TEXT           NOT NULL,
    price_date DATE           NOT NULL,
    open       NUMERIC(12, 4),
    high       NUMERIC(12, 4),
    low        NUMERIC(12, 4),
    close      NUMERIC(12, 4) NOT NULL,
    volume     BIGINT,
    PRIMARY KEY (ticker, price_date)
);

CREATE INDEX idx_historical_prices_ticker_date ON historical_prices(ticker, price_date DESC);
```

**Cache invalidation rule:** data is considered fresh if the most recent `price_date` equals yesterday (last completed trading day). Otherwise the Market Data Service refetches from AlphaVantage.

---

## portfolios *(Phase 4 — future)*

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `user_id` | UUID | Owner (from Portfolio/Auth service) |
| `name` | TEXT | Human-readable label |
| `created_at` | TIMESTAMPTZ | |

---

## portfolio_assets *(Phase 4 — future)*

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `portfolio_id` | UUID | FK → portfolios |
| `ticker` | TEXT | Asset symbol (e.g. `AAPL`) |
| `weight` | NUMERIC(6,4) | Allocation weight (0–1, sum = 1.0 per portfolio) |

---

## Tooling

- **sqlc** (`sqlc.yaml`) generates type-safe Go from SQL queries — run `make generate` after any schema or query change
- **pgx/v5** connection pool (`internal/db/db.go`)
- Migrations live in `internal/db/migrations/` (numbered `.up.sql` / `.down.sql`)
- Commands: `make migrate-up`, `make migrate-down 1`

---

## Related Notes

- [[Architecture]]
- [[Worker Service]]
- [[Risk Metrics]]
