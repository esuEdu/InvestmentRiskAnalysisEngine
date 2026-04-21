Generated with **sqlc** — queries live in `internal/db/generated/`.

---

## analysis_requests

Stores each incoming analysis job.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `portfolio_id` | UUID | Reference to the portfolio being analysed |
| `benchmark` | TEXT | Optional ticker symbol (e.g. `SPY`) |
| `period` | TEXT | Time window (e.g. `1y`, `6m`) |
| `status` | TEXT | `pending` / `processing` / `completed` / `failed` |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

---

## portfolios *(planned)*

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `user_id` | UUID | Owner |
| `name` | TEXT | Human-readable label |
| `created_at` | TIMESTAMPTZ | |

---

## portfolio_assets *(planned)*

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `portfolio_id` | UUID | FK → portfolios |
| `ticker` | TEXT | Asset symbol (e.g. `AAPL`) |
| `weight` | NUMERIC | Allocation weight (0–1) |

---

## analysis_results *(planned)*

| Column | Type | Notes |
|---|---|---|
| `analysis_request_id` | UUID | FK → analysis_requests |
| `annualized_volatility` | NUMERIC | |
| `sharpe_ratio` | NUMERIC | |
| `beta` | NUMERIC | |
| `max_drawdown` | NUMERIC | |
| `var_95` | NUMERIC | Value at Risk at 95% confidence |
| `concentration_score` | NUMERIC | Herfindahl index |
| `raw_metrics_json` | JSONB | Full metrics payload |

---

## historical_prices *(planned)*

| Column | Type | Notes |
|---|---|---|
| `ticker` | TEXT | Asset symbol |
| `price_date` | DATE | |
| `open` | NUMERIC | |
| `high` | NUMERIC | |
| `low` | NUMERIC | |
| `close` | NUMERIC | |
| `volume` | BIGINT | |

Primary key: `(ticker, price_date)`

---

## Tooling

- **sqlc** (`sqlc.yaml`) generates type-safe Go from SQL queries
- **pgx/v5** connection pool (`internal/db/db.go`)
- Migrations managed via scripts in `scripts/`

---

## Related Notes

- [[Architecture]]
- [[Risk Metrics]]
