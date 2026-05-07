## Current Status

Phase 3 (Risk Worker) complete. Now working on Phase 4 — Portfolio Service.

---

## Completed

- [x] Project scaffold (Clean Architecture layout)
- [x] PostgreSQL connection pool (`pgx/v5`)
- [x] sqlc code generation setup
- [x] `analysis_requests` table + migrations
- [x] `AnalysisRequest` domain model and status enum
- [x] Repository layer (Create, Get, List, UpdateStatus)
- [x] UseCase layer (ExecuteCreate, ExecuteGet, ExecuteList, ExecuteUpdate)
- [x] HTTP handler for `POST /api/v1/analyses` (returns 202 Accepted)
- [x] Gin router setup with `/api/v1` group
- [x] Structured logging with Uber Zap
- [x] Viper-based config loading
- [x] Fix `GET` / `List` route conflict
- [x] Implement `Get`, `List`, `Update` handlers
- [x] RabbitMQ connection and publisher (`internal/messaging`)
- [x] Publish job to `risk-analysis-jobs` on analysis creation
- [x] RabbitMQ consumer layer (`internal/messaging/consumer`)
- [x] DB migrations — `analysis_request_assets`, `analysis_results`, `historical_prices`
- [x] `internal/riskmetrics` — volatility, Sharpe, beta, drawdown, VaR, HHI calculator
- [x] `internal/marketdata` — AlphaVantage adapter, use case, repository, HTTP handler
- [x] `cmd/marketdata/main.go` — Market Data Service entrypoint (`:8081`)
- [x] `internal/worker` — `Handle` orchestration (fetch prices → compute → persist)
- [x] `cmd/worker/main.go` — Worker Service entrypoint with graceful shutdown
- [x] `GET /api/v1/analyses/:id/results` — results endpoint on API Service
- [x] API updated to accept multi-asset requests with portfolio weights
- [x] Config — `MarketDataServiceURL`, `MarketDataAPIKey`
- [x] Docker Compose — `worker` and `marketdata` services

> See [[Worker Service]] for implementation details.

### Phase 4 — Portfolio Service

- [ ] `portfolios` and `portfolio_assets` tables
- [ ] Portfolio CRUD endpoints
- [ ] Asset weight validation (sum must equal 1.0)

### Phase 5 — Infrastructure

- [ ] Kubernetes manifests (Deployments, Services, ConfigMaps, Secrets)
- [ ] NGINX Ingress resource
- [ ] Dockerfile for API and Worker
- [ ] GitHub Actions CI pipeline (lint, test, build)

### Phase 6 — Observability

- [ ] Prometheus metrics endpoint
- [ ] Grafana dashboards
- [ ] OpenTelemetry tracing

### Future Ideas

- [ ] Monte Carlo simulation
- [ ] Portfolio stress testing
- [ ] Sector exposure analysis
- [ ] User authentication (JWT)
- [ ] PDF risk report generation
- [ ] Webhook notifications on completion
- [ ] Frontend dashboard
- [ ] Kafka variant for event streaming

---

## Related Notes

- [[Architecture]]
- [[Risk Metrics]]
- [[API Reference]]
- [[Worker Service]]
