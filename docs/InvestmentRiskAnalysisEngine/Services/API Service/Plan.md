---
service: API Service (`cmd/api`)
port: `:8080`
status: active
---

# API Service — Plan

## Current Capabilities

- `POST /api/v1/analyses` — create analysis request (multi-asset, async, 202 Accepted)
- `GET /api/v1/analyses/:id` — fetch request status and metadata
- `GET /api/v1/analyses/:id/results` — fetch computed risk metrics
- `GET /api/v1/analyses` — list requests with pagination and status filter
- `PUT /api/v1/analyses/:id` — update status (internal/admin)
- `GET /api/v1/health` — liveness probe

---

## Short-term

- [ ] **Input validation hardening** — validate ticker format, period enum, weight precision limits
- [ ] **Idempotency key** — allow clients to safely retry `POST /analyses` without duplicating jobs
- [ ] **Bruno collection** — cover `GET /analyses/:id/results` in the existing collection

---

## Medium-term

- [ ] **Authentication (JWT)** — protect all endpoints; integrate with the future Portfolio Service for user identity
- [ ] **Per-user scoping** — `GET /analyses` should only return analyses owned by the authenticated user
- [ ] **Rate limiting** — per-IP or per-user request cap on `POST /analyses` to guard the AlphaVantage quota
- [ ] **Webhook on completion** — optional `callback_url` on the create request; Worker notifies on finish

---

## Long-term

- [ ] **API versioning** — `v2` path when breaking changes are needed
- [ ] **WebSocket status stream** — `WS /api/v1/analyses/:id/status` for real-time polling replacement
- [ ] **Batch create** — `POST /api/v1/analyses/batch` to queue multiple portfolios atomically
- [ ] **PDF report endpoint** — `GET /api/v1/analyses/:id/report` generates a risk summary PDF

---

## Related Notes

- [[API Reference]]
- [[Architecture]]
- [[Project Plan]]
