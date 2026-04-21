## Prerequisites

- Go 1.26+
- Docker + Docker Compose
- `sqlc` CLI (for regenerating DB code)
- `make` (optional but helpful)

---

## Getting Started

### 1. Clone and configure

```bash
git clone https://github.com/esuEdu/investment-risk-engine
cd investment-risk-engine
```

Copy and edit the environment config (via Viper / `.env` or env vars):

| Variable | Description |
|---|---|
| `APP_ENV` | `development` / `production` |
| `DB_HOST` | PostgreSQL host |
| `DB_PORT` | PostgreSQL port (default `5432`) |
| `DB_USER` | Database user |
| `DB_PASSWORD` | Database password |
| `DB_NAME` | Database name |

### 2. Start infrastructure

```bash
docker compose up -d postgres redis rabbitmq
```

### 3. Run the API

```bash
go run ./cmd/api
```

The server starts on `:8080`.

---

## Project Structure

```
cmd/
  api/main.go            — Wires dependencies and starts the server

internal/
  config/                — Viper config loader
  db/                    — pgx pool + sqlc generated queries
  server/                — Gin engine, middleware, routers, response helpers
  analysis/
    delivery/http/       — Gin handlers
    usecase/             — Business logic
    repository/          — DB access (sqlc)
    domain/              — Models + interfaces

pkg/
  logger/                — Zap logger

deployments/             — Docker and Kubernetes manifests
scripts/                 — DB migrations and helper scripts
docs/                    — This Obsidian vault
```

---

## Regenerating DB Code

After modifying SQL queries:

```bash
sqlc generate
```

Config: `sqlc.yaml`

---

## Useful Make Targets

```bash
make run       # Run the API locally
make test      # Run all tests
make build     # Build the binary
```

---

## Branching Strategy

| Branch prefix | Purpose |
|---|---|
| `feature/<issue>-<name>` | New features |
| `fix/<issue>-<name>` | Bug fixes |
| `chore/…` | Tooling, config, docs |

PRs target `main`.

---

## Related Notes

- [[Architecture]]
- [[Infrastructure]]
