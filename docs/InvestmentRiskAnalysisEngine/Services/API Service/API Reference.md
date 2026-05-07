# API Reference

Base path: `/api/v1`

---

## Health

### `GET /api/v1/health`

Returns service liveness status.

**Response `200 OK`**
```json
{ "service": "api" }
```

---

## Analyses

### `POST /api/v1/analyses`

Create a new analysis request. Job is queued asynchronously.

**Request body**
```json
{
  "benchmark": "SPY",
  "period": "1y",
  "assets": [
    { "ticker": "AAPL", "weight": 0.40 },
    { "ticker": "MSFT", "weight": 0.35 },
    { "ticker": "GOOGL", "weight": 0.25 }
  ]
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `benchmark` | string | no | Ticker used as market benchmark (e.g. `"SPY"`). Required for Beta calculation. |
| `period` | string | yes | Time window: `"1y"`, `"6m"`, `"3m"` |
| `assets` | array | yes | At least one asset required |
| `assets[].ticker` | string | yes | Asset ticker symbol (e.g. `"AAPL"`) |
| `assets[].weight` | number | yes | Portfolio weight. All weights must sum to exactly `1.0`. |

**Validation errors:**
- Weights do not sum to `1.0` → `400 Bad Request`
- `assets` is empty → `400 Bad Request`
- Unknown `period` value → `400 Bad Request`

**Response `202 Accepted`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "benchmark": "SPY",
  "period": "1y",
  "assets": [
    { "ticker": "AAPL", "weight": 0.40 },
    { "ticker": "MSFT", "weight": 0.35 },
    { "ticker": "GOOGL", "weight": 0.25 }
  ],
  "created_at": "2026-05-06T10:00:00Z",
  "updated_at": "2026-05-06T10:00:00Z"
}
```

> `202 Accepted` means the job was queued, not yet processed. Poll `GET /analyses/:id` to track status.

---

### `GET /api/v1/analyses/:id`

Fetch a single analysis request by ID.

**Response `200 OK`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "benchmark": "SPY",
  "period": "1y",
  "assets": [
    { "ticker": "AAPL", "weight": 0.40 },
    { "ticker": "MSFT", "weight": 0.35 },
    { "ticker": "GOOGL", "weight": 0.25 }
  ],
  "created_at": "2026-05-06T10:00:00Z",
  "updated_at": "2026-05-06T10:05:00Z"
}
```

---

### `GET /api/v1/analyses`

List analysis requests with optional filtering.

**Query params**

| Param | Type | Notes |
|---|---|---|
| `limit` | int | Max results to return (default: 20) |
| `offset` | int | Pagination offset (default: 0) |
| `status` | string | Filter by status: `pending`, `processing`, `completed`, `failed` |

**Response `200 OK`**
```json
{
  "data": [ ... ],
  "meta": { "limit": 20, "offset": 0, "total": 42 }
}
```

---

### `GET /api/v1/analyses/:id/results`

Fetch the computed risk metrics for a completed analysis.

Returns `404 Not Found` if the analysis is still `pending` or `processing`.

**Response `200 OK`**
```json
{
  "analysis_request_id": "550e8400-e29b-41d4-a716-446655440000",
  "annualized_volatility": 0.182341,
  "sharpe_ratio": 1.243500,
  "beta": 1.120000,
  "max_drawdown": -0.213400,
  "var_95": -0.028700,
  "concentration_score": 0.355000,
  "created_at": "2026-05-06T10:05:00Z"
}
```

| Field | Notes |
|---|---|
| `annualized_volatility` | Standard deviation of returns × √252 |
| `sharpe_ratio` | Risk-adjusted return (rf = 5%) |
| `beta` | Sensitivity vs benchmark. `null` if no benchmark was provided. |
| `max_drawdown` | Largest peak-to-trough loss. Always negative. |
| `var_95` | Worst expected daily loss at 95% confidence. Always negative. |
| `concentration_score` | Herfindahl index Σ w_i². Range: 0 (diversified) → 1 (single asset). |

---

### `PUT /api/v1/analyses/:id`

Update an analysis status (admin / internal use).

**Request body**
```json
{ "status": "failed" }
```

---

## Status Values

| Status | Meaning |
|---|---|
| `pending` | Request created, not yet picked up by Worker |
| `processing` | Worker is fetching prices and calculating metrics |
| `completed` | Results available — call `GET /analyses/:id/results` |
| `failed` | Calculation failed (bad ticker, quota exceeded, etc.) |

---

## Error Responses

| Code | Scenario |
|---|---|
| `400 Bad Request` | Invalid or missing fields, weights do not sum to 1.0 |
| `404 Not Found` | Analysis or results not found |
| `500 Internal Server Error` | Failed to persist or queue the analysis |

---

## Related Notes

- [[Architecture]]
- [[Worker Service]]
- [[Risk Metrics]]
