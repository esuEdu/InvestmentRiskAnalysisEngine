---
service: Worker Service (`cmd/worker`)
port: none (RabbitMQ consumer)
status: active
---

# Worker Service — Plan

## Current Capabilities

- Consumes `risk-analysis-jobs` queue from RabbitMQ
- Fetches historical prices for each asset via Market Data Service HTTP
- Computes 6 risk metrics via `internal/riskmetrics`: volatility, Sharpe, beta, max drawdown, VaR95, HHI
- Persists results to `analysis_results` and updates `analysis_requests.status`
- Marks jobs `failed` on any error; graceful shutdown on SIGTERM

---

## Short-term

- [ ] **Worker unit tests** — mock `MarketDataClient` and `AnalysisRepository` to test the `Handle` flow end-to-end
- [ ] **Dead-letter queue (DLQ)** — configure RabbitMQ DLQ so failed jobs are requeued up to N times before being parked
- [ ] **Retry backoff** — distinguish transient errors (5xx, timeout) from permanent ones (bad ticker, quota) before requeuing

---

## Medium-term

- [ ] **Parallel price fetching** — use `errgroup` to fetch all asset prices concurrently instead of sequentially
- [ ] **Correlation matrix** — add pairwise return correlation to `MetricsResult` and `analysis_results`
- [ ] **Benchmark optional warning** — log a structured warning when Beta is skipped (no benchmark) rather than silently returning `null`
- [ ] **Observability** — emit Prometheus counters for jobs processed, failed, duration histogram

---

## Long-term

- [ ] **Monte Carlo simulation** — extend `riskmetrics` with simulated return paths for probabilistic risk estimation
- [ ] **Stress testing** — scenario-based shocks (e.g. -30% market crash) applied to the portfolio
- [ ] **Sector exposure** — tag each ticker with its sector and compute sector concentration alongside HHI
- [ ] **Notification publish** — after completing a job, publish to a `notifications` queue for the API/Portfolio Service to notify the user

---

## Related Notes

- [[Worker Service]]
- [[Architecture]]
- [[Risk Metrics]]
- [[Project Plan]]
