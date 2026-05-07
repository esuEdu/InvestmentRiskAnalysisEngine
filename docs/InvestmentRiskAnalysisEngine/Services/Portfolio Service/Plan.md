---
service: Portfolio Service (`cmd/portfolio`)
port: `:8082`
status: planned (Phase 4)
---

# Portfolio Service — Plan

## Purpose

Owns user identity, portfolio management, and asset allocation. Other services reference a `portfolio_id` but cannot mutate portfolio data — all writes go through this service.

Isolated for security: JWT issuance, session management, and user PII stay in one boundary.

---

## Planned Database Tables

```sql
-- User accounts
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Named portfolios belonging to a user
CREATE TABLE portfolios (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Assets and weights within a portfolio (sum must equal 1.0)
CREATE TABLE portfolio_assets (
    portfolio_id UUID NOT NULL REFERENCES portfolios(id),
    ticker       TEXT NOT NULL,
    weight       NUMERIC(6,4) NOT NULL CHECK (weight > 0 AND weight <= 1),
    PRIMARY KEY (portfolio_id, ticker)
);
```

---

## Planned Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/auth/register` | Create user account |
| `POST` | `/api/v1/auth/login` | Issue JWT |
| `GET` | `/api/v1/portfolios` | List user's portfolios |
| `POST` | `/api/v1/portfolios` | Create a portfolio with assets |
| `GET` | `/api/v1/portfolios/:id` | Fetch portfolio detail |
| `PUT` | `/api/v1/portfolios/:id` | Update name or asset weights |
| `DELETE` | `/api/v1/portfolios/:id` | Delete portfolio |
| `POST` | `/api/v1/portfolios/:id/analyses` | Trigger analysis for this portfolio |
| `GET` | `/api/v1/portfolios/:id/risk-summary` | Aggregate latest analysis results |

---

## Short-term (when starting this service)

- [ ] DB migrations — `users`, `portfolios`, `portfolio_assets`
- [ ] JWT middleware (sign + verify with shared secret or RS256 key pair)
- [ ] User registration and login endpoints
- [ ] Portfolio CRUD with weight-sum validation (must equal 1.0)
- [ ] `POST /portfolios/:id/analyses` — calls API Service internally to trigger a job

---

## Medium-term

- [ ] Refresh tokens — short-lived access token + long-lived refresh token rotation
- [ ] Ownership enforcement — all portfolio endpoints verify the JWT user owns the portfolio
- [ ] `GET /portfolios/:id/risk-summary` — aggregates the latest `analysis_results` for the portfolio's assets
- [ ] Portfolio history — track changes to asset weights over time

---

## Long-term

- [ ] Social sharing — public portfolio view with a shareable link
- [ ] Multi-currency support — normalize weights by currency before analysis
- [ ] Import from broker — parse CSV exports from common brokers (IBKR, Schwab) into portfolio assets

---

## Related Notes

- [[Architecture]]
- [[Database Schema]]
- [[Project Plan]]
