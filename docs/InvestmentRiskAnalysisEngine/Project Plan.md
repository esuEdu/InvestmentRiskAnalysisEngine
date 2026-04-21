## Current Status

The project is in active development on the `feature/7-ImplementAnalysisEndpoint` branch.

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

---

## In Progress

- [x] Fix `GET` / `List` route conflict (both mapped to `GET ""`)
- [x] Implement `Get` handler (fetch single analysis by ID from path param)
- [x] Implement `List` handler (with `limit`, `offset`, `status` query params)
- [x] Implement `Update` handler (status transition)

---

## Backlog

### Phase 2 — Messaging

- [ ] RabbitMQ connection and publisher
- [ ] Publish job to `risk-analysis-jobs` on analysis creation
- [ ] Dead-letter queue configuration

### Phase 3 — Risk Worker

- [ ] Worker service entrypoint (`cmd/worker`)
- [ ] RabbitMQ consumer
- [ ] Market data client (AlphaVantage / TwelveData / Polygon)
- [ ] Historical price storage (`historical_prices` table)
- [ ] Risk metric calculators:
  - [ ] Annualized volatility
  - [ ] Sharpe ratio
  - [ ] Beta
  - [ ] Maximum drawdown
  - [ ] Historical VaR (95%)
  - [ ] Concentration score (HHI)
- [ ] Persist results to `analysis_results`
- [ ] Update `analysis_requests.status` → `completed` / `failed`

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
