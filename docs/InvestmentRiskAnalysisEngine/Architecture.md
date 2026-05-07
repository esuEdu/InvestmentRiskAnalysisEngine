# Architecture

---

## Services vs Internal Packages

This system is built with **multiple independent services** and **shared internal libraries**.

### Services (separate processes, own binary)

| Service | Binary | Port | What it is |
|---|---|---|---|
| **API** | `cmd/api` | `:8080` | HTTP entry point for clients |
| **Worker** | `cmd/worker` | — | Background job processor (no HTTP) |
| **Market Data** | `cmd/marketdata` | `:8081` | Fetches + stores + serves historical prices |
| **Portfolio** | `cmd/portfolio` *(future)* | `:8082` | User profiles, portfolios, asset weights |

Each service runs as its own process, scales independently, and owns its domain.

### Internal Packages (libraries, NOT processes)

| Package | Imported by | What it is |
|---|---|---|
| `internal/riskmetrics` | Worker | Pure math functions — volatility, Sharpe, beta, VaR, etc. |
| `internal/analysis` | API + Worker | Domain model, repository, use cases for analysis requests |
| `internal/messaging` | API + Worker | RabbitMQ connection, publisher, consumer |
| `internal/marketdata` | Market Data Service | Provider adapters, use cases, repository for prices |
| `internal/config` | All services | Viper config loader |
| `internal/db` | All services | pgx pool factory + sqlc-generated queries |
| `pkg/logger` | All services | Uber Zap singleton |

These are **Go packages compiled into service binaries** — not processes, not HTTP servers, not independently deployable.

---

## High-Level Flow

```
Client
  │
  ▼
NGINX Ingress
  │
  ▼
API Service  (:8080)
  │  validates request
  │  saves AnalysisRequest → PostgreSQL
  └──publishes job ──────────────────────────────────┐
                                                      ▼
                                                 RabbitMQ
                                        "risk-analysis-jobs"
                                                      │
                                                      ▼
                                            Worker Service
                                              │  consumes job
                                              │  HTTP call
                                              ▼
                                    Market Data Service (:8081)
                                      │  checks historical_prices in DB
                                      │  if stale → fetches from AlphaVantage/TwelveData
                                      └──returns []PricePoint to Worker
                                              │
                                              │  riskmetrics.Compute() [internal package]
                                              │
                                              ▼
                                           PostgreSQL
                                    analysis_results + status update

                                    Portfolio Service (:8082)  [future]
                                      user profiles, portfolios, asset weights
                                      other services reference portfolio_id
```

---

## API Service (`cmd/api`)

Follows **Clean Architecture** (Delivery → Use Case → Repository → Domain).

```
internal/analysis/
  delivery/http/   — Gin handlers
  usecase/         — Business logic
  repository/      — PostgreSQL queries (sqlc)
  domain/          — AnalysisRequest model + interfaces
```

Wiring order:
1. Load config → connect PostgreSQL → connect RabbitMQ
2. Build Repository → UseCase → Handler chain
3. Start Gin on `:8080`

---

## Worker Service (`cmd/worker`)

Background service. **Has no HTTP server.**

```
internal/worker/
  worker.go        — Handler: orchestrates steps 1–5 per job
```

Imports:
- `internal/riskmetrics` — pure math (internal package)
- `internal/messaging/consumer` — RabbitMQ consumer (internal package)
- `internal/analysis/repository` — UpdateStatus, CreateResult (internal package)
- Calls **Market Data Service over HTTP** (separate process)

Wiring order:
1. Load config → connect PostgreSQL → connect RabbitMQ
2. Build repository + HTTP client for Market Data Service
3. Build `riskmetrics.Calculator` (in-process)
4. Start `AnalysisConsumer.Start(ctx, worker.Handle)` — blocks until shutdown

---

## Market Data Service (`cmd/marketdata`)

**Own binary, own HTTP server on `:8081`.**

Why it's a service and not just a package:
- Needs to run continuously and refresh stale prices on a schedule
- Future: serves chart/graph data directly to clients or frontend
- Has its own DB table (`historical_prices`) — other services must not query it directly

```
internal/marketdata/
  domain/
    model.go         — PricePoint, OHLCV structs
    provider.go      — Provider interface
    repository.go    — Repository interface
  usecase/
    usecase.go       — GetPrices(ticker, period): check DB → fetch if stale → return
  repository/
    mapper.go        — DB row → domain type conversion
    repository.go    — reads/writes historical_prices
  delivery/http/
    handler.go       — GET /prices?ticker=AAPL&period=1y
  provider/
    alphaVantage.go  — AlphaVantage adapter + rate limiter (5 req/min)
```

Current endpoints:
- `GET /prices?ticker={ticker}&period={period}` → returns `[]PricePoint` (internal use)

Future endpoints:
- `GET /prices/:ticker/chart` → OHLCV data for charting
- `GET /prices/:ticker/latest` → latest close price
- WebSocket for real-time price streaming

---

## Portfolio Service (`cmd/portfolio`) *(future)*

**Own binary, own HTTP server on `:8082`.**

Why it's a service and not just a package:
- Will own **user profiles and authentication** (JWT, sessions) — security-sensitive, must be isolated
- Other services reference a `portfolio_id` but cannot mutate portfolio data
- Can evolve independently (e.g. add social features, sharing)

Planned responsibilities:
- User profile CRUD
- Portfolio CRUD (`portfolios` table)
- Asset weight management (`portfolio_assets` table, sum must = 1.0)
- `GET /portfolios/:id/risk-summary` — aggregate from analysis results

---

## Messaging Layer (RabbitMQ)

Decouples the API from heavy computation. Dead-letter queues handle retries.

| Queue | Publisher | Consumer | Purpose |
|---|---|---|---|
| `risk-analysis-jobs` | API Service | Worker Service | Trigger risk calculation |
| `market-data-refresh-jobs` *(future)* | Worker | Market Data Service | Refresh stale price data |
| `notifications` *(future)* | Worker | Portfolio/API | Notify user on completion |

---

## Package Map

```
cmd/
  api/main.go              — API service entrypoint
  worker/main.go           — Worker service entrypoint
  marketdata/main.go       — Market Data service entrypoint
  portfolio/main.go        — Portfolio service entrypoint (future)

internal/
  config/                  — Viper config loader (shared)
  db/                      — pgx pool + sqlc generated code (shared)
  server/                  — Gin engine + middleware helpers (shared)
  messaging/               — RabbitMQ connection, publisher, consumer (shared)
  analysis/                — Analysis domain: model, repo, usecase, HTTP handler
  marketdata/              — Market Data domain: model, repo, usecase, HTTP handler, providers
  riskmetrics/             — Pure math calculators (no DB, no HTTP)
  worker/                  — Worker orchestration handler

pkg/
  logger/                  — Uber Zap singleton (shared)
  middleware/              — Zap request-logging Gin middleware (shared)
  response/                — Standardized REST response envelope (shared)
```

---

## Key Design Decisions

- **sqlc** for type-safe SQL — no ORM, full query control.
- **pgx/v5** connection pool for all services.
- **Clean Architecture** per service — domain logic stays independent of HTTP and DB.
- **202 Accepted** on `POST /analyses` — creation is async; clients poll for results.
- **Services communicate over HTTP or RabbitMQ** — never import each other's internal packages.
- **Internal packages are shared libraries** — `internal/riskmetrics` is compiled into the Worker, not deployed separately.
- **Market Data owns its table** — Worker calls Market Data Service over HTTP, never queries `historical_prices` directly.

---

## Related Notes

- [[API Reference]]
- [[Database Schema]]
- [[Infrastructure]]
- [[Worker Service]]
- [[Project Plan]]
