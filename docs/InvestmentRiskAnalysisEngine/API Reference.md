Base path: `/api/v1`

---

## Health

### `GET /api/v1/health`

Returns service liveness status.

**Response `200 OK`**
```json
{
  "service": "api"
}
```

---

## Analyses

### `POST /api/v1/analyses`

Create a new analysis request. The job is queued asynchronously.

**Request body**
```json
{
  "portfolio_id": "uuid",
  "benchmark": "SPY",
  "period": "1y"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `portfolio_id` | UUID string | yes | Must be a valid UUID |
| `benchmark` | string | no | Ticker symbol (e.g. `"SPY"`) |
| `period` | string | yes | e.g. `"1y"`, `"6m"`, `"3m"` |

**Response `202 Accepted`**
```json
{
  "id": "uuid",
  "status": "pending",
  "benchmark": "SPY",
  "period": "1y",
  "created_at": "2026-04-21T00:00:00Z",
  "updated_at": "2026-04-21T00:00:00Z"
}
```

> `202 Accepted` signals that the request was received and queued, not yet processed.

---

### `GET /api/v1/analyses`

Get a single analysis or list analyses.

> **Note:** The `GET` and `List` routes currently share the same path — route disambiguation is in progress.

---

### `PUT /api/v1/analyses`

Update an analysis (e.g. update status after worker completes).

---

## Status Values

| Status | Meaning |
|---|---|
| `pending` | Request created, not yet picked up by worker |
| `processing` | Worker is calculating metrics |
| `completed` | Results are available |
| `failed` | Calculation failed |

---

## Error Responses

| Code | Scenario |
|---|---|
| `400 Bad Request` | Invalid or missing fields in request body |
| `500 Internal Server Error` | Failed to persist or queue the analysis |

---

## Related Notes

- [[Architecture]]
- [[Risk Metrics]]
