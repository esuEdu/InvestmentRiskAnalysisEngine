---
service: Market Data Service (`cmd/marketdata`)
port: `:8081`
status: active
---

# Market Data Service — Plan

## Current Capabilities

- `GET /api/v1/prices?ticker=X&period=Y` — cache-first price fetch (AlphaVantage → PostgreSQL)
- Single provider: AlphaVantage Free Plan (5 req/min, 25 req/day)
- Token-bucket rate limiter on all outbound calls
- Returns daily closing prices filtered to the requested period

---

## Short-term

- [ ] **Period validation** — return `400 Bad Request` for unsupported `period` values instead of silently returning empty data
- [ ] **Market-closed awareness** — skip today when computing "yesterday" on weekends and public holidays
- [ ] **Bruno collection** — add `GET /prices` requests to the Bruno API collection for manual testing
- [ ] **Health endpoint improvement** — include DB connectivity check in the health response

---

## Medium-term

- [ ] **TwelveData fallback provider** — implement a second adapter; if AlphaVantage quota is hit, fall back automatically
- [ ] **Provider failover interface** — abstract provider selection behind a `CompositeProvider` that tries providers in order
- [ ] **Daily quota guard** — track daily call count in Redis or a DB table; return a clear error when the limit is reached rather than propagating a 429 from AlphaVantage
- [ ] **Scheduled refresh** — cron job (or RabbitMQ message) to pre-warm price cache for frequently requested tickers nightly
- [ ] **OHLCV chart endpoint** — `GET /prices/:ticker/chart` returning full OHLCV for frontend charting
- [ ] **Latest price endpoint** — `GET /prices/:ticker/latest` returning only the most recent close

---

## Long-term

- [ ] **Real-time prices (WebSocket)** — `WS /prices/:ticker/stream` for live price streaming to a future frontend
- [ ] **Additional data providers** — Yahoo Finance, Polygon.io, or Alpha Vantage Premium for higher quotas
- [ ] **Dividend and split adjustment** — store split/dividend-adjusted close prices separately
- [ ] **Intraday data** — support `1d` period with minute-level granularity for short-horizon analysis

---

## Related Notes

- [[Market Data Service]]
- [[Architecture]]
- [[Database Schema]]
- [[Project Plan]]
