---
service: Market Data Service (`cmd/marketdata`)
port: `:8081`
status: active (internal use only)
---

# Market Data Service

Internal HTTP service that provides historical price data to the Worker Service. Not exposed to external clients.

---

## Responsibilities

- Serve daily closing prices for a given ticker and time period
- Cache-first: check `historical_prices` in PostgreSQL before calling the external provider
- Fetch from AlphaVantage when data is missing or stale (most recent date < yesterday)
- Rate-limit outbound AlphaVantage calls: 5 req/min (one token every 12 seconds via token bucket)

---

## Endpoint

```
GET /api/v1/prices?ticker={ticker}&period={period}
```

| Param | Values | Notes |
|---|---|---|
| `ticker` | e.g. `AAPL`, `SPY` | Case-insensitive |
| `period` | `1y`, `6m`, `3m`, `1m` | Translated to a date cutoff |

**Response `200 OK`**
```json
{
  "data": [
    { "date": "2025-05-01", "close": 183.25 },
    ...
  ]
}
```

**Health**
```
GET /health → { "service": "marketdata" }
```

---

## Architecture

```
internal/marketdata/
  domain/
    model.go         — PricePoint, OHLCV structs
    provider.go      — Provider interface (FetchDailyPrices)
    repository.go    — Repository interface (GetPrices, GetLatestPriceDate, UpsertPrices)
  usecase/
    usecase.go       — freshness check → fetch if stale → filter by period → return
  repository/
    mapper.go        — pgtype conversions
    repository.go    — reads/writes historical_prices table
  delivery/http/
    handler.go       — GET /api/v1/prices
  provider/
    alphaVantage.go  — AlphaVantage adapter + token-bucket rate limiter
```

---

## Cache Freshness Rule

Data is considered fresh if `latest price_date ≥ yesterday` (last Friday if today is Monday). Otherwise the provider is called and the result is upserted into `historical_prices`.

---

## AlphaVantage Free Plan Limits

| Limit | Value |
|---|---|
| Requests per minute | 5 |
| Requests per day | 25 |

A 5-asset portfolio uses 5 API calls if no data is cached. The first real run after a cold start will exhaust a large portion of the daily quota.

---

## Related Notes

- [[Architecture]]
- [[Database Schema]]
- [[Plan]]
