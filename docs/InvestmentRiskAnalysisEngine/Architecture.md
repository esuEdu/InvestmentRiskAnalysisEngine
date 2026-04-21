## High-Level Flow

```
Client
  │
  ▼
NGINX Ingress
  │
  ▼
API Service (Go / Gin)
  │  ├── Validates request
  │  ├── Persists AnalysisRequest → PostgreSQL
  │  └── Publishes job → RabbitMQ
  ▼
RabbitMQ
  │
  ▼
Risk Worker (Go)          ←── Market Data Service
  │  ├── Consumes job
  │  ├── Fetches historical prices
  │  ├── Calculates risk metrics
  │  └── Persists results → PostgreSQL
  ▼
PostgreSQL / Redis
```

---

## Layer Breakdown

### API Service

Follows **Clean Architecture** (Delivery → Use Case → Repository → Domain).

```
internal/analysis/
  ├── delivery/http/   — Gin handlers (HTTP layer)
  ├── usecase/         — Business logic
  ├── repository/      — PostgreSQL queries (sqlc-generated)
  └── domain/          — Models and interfaces
```

Entry point: `cmd/api/main.go`

Wiring order:
1. Load config (`internal/config`)
2. Connect PostgreSQL pool (`internal/db`)
3. Instantiate sqlc Queries
4. Build Repository → UseCase → Handler chain
5. Start Gin server on `:8080`

---

### Risk Worker *(planned)*

Background service that:
1. Consumes jobs from the `risk-analysis-jobs` queue
2. Fetches historical market data from an external provider
3. Computes risk metrics (volatility, Sharpe, beta, VaR, etc.)
4. Writes results back to PostgreSQL

---

### Market Data Service *(planned)*

Responsible for:
- Fetching OHLCV time-series from external APIs
- Normalising and storing historical prices
- Refreshing stale data on demand

Candidate providers: AlphaVantage, TwelveData, Polygon, Yahoo Finance.

---

### Messaging Layer (RabbitMQ)

Decouples the API from heavy computation.

| Queue | Purpose |
|---|---|
| `risk-analysis-jobs` | Trigger risk calculation for a request |
| `market-data-refresh-jobs` | Refresh stale price data |
| `notifications` | Async user notifications |

Dead-letter queues handle failed jobs with automatic retries.

---

## Package Map

```
cmd/
  api/main.go              — API entrypoint

internal/
  config/                  — Viper-based config loader
  db/                      — pgx pool factory + sqlc generated code
  server/                  — Gin engine, middleware, response helpers, routers
  analysis/
    delivery/http/         — HTTP handlers
    usecase/               — Use cases (ExecuteCreate, ExecuteGet, …)
    repository/            — sqlc repository + mapper
    domain/                — AnalysisRequest model, Repository/Queue interfaces

pkg/
  logger/                  — Uber Zap logger singleton
```

---

## Key Design Decisions

- **sqlc** for type-safe SQL — no ORM, full control over queries.
- **pgx/v5** connection pool for PostgreSQL.
- **Clean Architecture** layers keep domain logic independent of HTTP and DB.
- **202 Accepted** on `POST /analyses` — creation is async; the client polls for results.
- **Uber Zap** for structured, high-performance logging.

---

## Related Notes

- [[API Reference]]
- [[Database Schema]]
- [[Infrastructure]]
