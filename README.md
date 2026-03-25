# Investment Risk Analysis Engine

A backend-focused **portfolio risk analysis platform** built with **Golang**, designed to analyze investment portfolios using historical market data and asynchronous processing.

This project is aimed at backend learning with strong emphasis on:

- Go service design
- clean architecture
- event-driven workflows
- background processing
- financial calculations
- containerization
- Kubernetes deployment
- messaging with Kafka or RabbitMQ

---

## Goal

Build a system where a user can submit a portfolio and receive a risk analysis report containing metrics such as:

- annualized volatility
- Sharpe ratio
- beta
- maximum drawdown
- Value at Risk (VaR 95%)
- concentration risk
- correlation matrix

This is not just a CRUD app. The main value is in the **backend architecture**, **financial domain logic**, and **distributed processing**.

---

## Why this project is strong

This project helps demonstrate real backend skills:

- API design with Go
- asynchronous job processing
- domain-driven modeling
- resilient communication between services
- data ingestion pipelines
- caching
- observability
- scalable deployment on Kubernetes

It is a good portfolio project for a **mid-level backend developer** who wants to move toward **senior backend/system design** level.

---

## Proposed Architecture

You can build it as a **modular monolith first**, then split parts into services if you want.

### Main components

1. **API Service**
    - receives portfolio requests
    - validates data
    - creates analysis jobs
    - exposes report endpoints

2. **Market Data Service**
    - fetches historical prices from external providers
    - normalizes and stores time-series data
    - refreshes stale data

3. **Risk Engine Worker**
    - consumes analysis jobs
    - calculates portfolio metrics
    - stores finished reports

4. **Notification Service** _(optional later)_
    - sends email/webhook when report is ready

5. **Cache Layer**
    - stores recent reports and price lookups

6. **Database**
    - stores portfolios, analyses, results, historical prices

---

## Suggested Stack

### Core

- **Golang**
- **Gin** or **Fiber** for HTTP API
- **PostgreSQL**
- **Redis**

### Messaging

Choose one:

#### Option A: Kafka

Best if your goal is:

- event-driven architecture
- stream-style thinking
- scalable processing
- stronger distributed systems experience

#### Option B: RabbitMQ

Best if your goal is:

- simpler queue setup
- worker/job processing
- easier local development
- reliable task-based workflows

### Infrastructure

- **Docker**
- **Kubernetes**
- **Helm** _(optional)_
- **Prometheus + Grafana** _(optional later)_

### Go Libraries

Possible choices:

- HTTP: `gin-gonic/gin` or `gofiber/fiber`
- ORM / DB: `gorm` or `sqlc` + `pgx`
- Validation: `go-playground/validator`
- Messaging:
    - Kafka: `segmentio/kafka-go` or `confluent-kafka-go`
    - RabbitMQ: `amqp091-go`
- Config: `viper`
- Logging: `zap` or `zerolog`

---

## High-Level Flow

```text
Client
  |
  v
API Service
  |
  | 1. validate portfolio
  | 2. create analysis record
  | 3. publish analysis job
  v
Kafka / RabbitMQ
  |
  v
Risk Engine Worker
  |
  | 4. fetch market data
  | 5. calculate metrics
  | 6. save result
  v
PostgreSQL / Redis
  |
  v
API Service
  |
  v
Client fetches report
```

---

## Core Use Case

A client sends a portfolio like:

```json
{
	"assets": [
		{ "ticker": "AAPL", "weight": 0.4 },
		{ "ticker": "MSFT", "weight": 0.3 },
		{ "ticker": "SPY", "weight": 0.2 },
		{ "ticker": "BTC-USD", "weight": 0.1 }
	],
	"period": "1y",
	"benchmark": "SPY"
}
```

The system returns a report like:

```json
{
	"analysisId": "an_123",
	"status": "completed",
	"result": {
		"annualizedVolatility": 0.22,
		"sharpeRatio": 1.08,
		"beta": 1.11,
		"maxDrawdown": -0.18,
		"historicalVaR95": -0.027,
		"concentrationScore": 0.3,
		"riskLevel": "moderate_high"
	}
}
```

---

## Main Features

### MVP

- create portfolio analysis request
- fetch historical prices
- compute volatility
- compute Sharpe ratio
- compute beta against benchmark
- compute maximum drawdown
- compute historical VaR 95%
- compute concentration score
- persist analysis results
- asynchronous processing through queue

### Phase 2

- correlation matrix
- sector exposure
- multi-period analysis
- report history
- authentication

### Phase 3

- stress testing
- Monte Carlo simulation
- alerting
- downloadable PDF/CSV reports
- dashboard frontend

---

## Recommended Service Breakdown

### 1. API Service

Responsibilities:

- expose REST endpoints
- validate requests
- persist portfolio and job metadata
- publish analysis events
- serve analysis reports

Example endpoints:

```http
POST /api/v1/analyses
GET /api/v1/analyses/{id}
GET /api/v1/portfolios/{id}/analyses
GET /health
```

### 2. Market Data Service

Responsibilities:

- fetch data from external APIs
- store OHLCV history
- refresh stale data
- deduplicate and normalize symbols

Possible external sources:

- Alpha Vantage
- Twelve Data
- Yahoo Finance wrappers
- Polygon
- Finnhub

### 3. Risk Worker

Responsibilities:

- consume jobs from Kafka or RabbitMQ
- load price history
- align time-series
- compute returns
- calculate metrics
- save final report

### 4. Notification Worker _(optional)_

Responsibilities:

- react when analysis status changes to `completed`
- send webhook/email/Slack notification

---

## Domain Model

### Portfolio

Represents a collection of assets and weights.

### Asset Allocation

Represents ticker, asset type, and weight.

### Analysis Request

Represents a snapshot of a requested analysis.

### Analysis Result

Represents the output metrics for a request.

### Historical Price

Represents time-series market data per ticker.

---

## Database Design

### portfolios

- id
- user_id
- name
- created_at
- updated_at

### portfolio_assets

- id
- portfolio_id
- ticker
- weight
- asset_type

### analysis_requests

- id
- portfolio_id
- benchmark
- period
- status
- created_at
- updated_at
- requested_by

### analysis_results

- id
- analysis_request_id
- annualized_volatility
- sharpe_ratio
- beta
- max_drawdown
- var_95
- concentration_score
- correlation_matrix_json
- raw_metrics_json
- created_at

### historical_prices

- id
- ticker
- price_date
- open
- high
- low
- close
- adjusted_close
- volume
- source
- created_at

### market_data_refresh_jobs

- id
- ticker
- status
- last_attempt_at
- error_message

---

## Key Financial Metrics

### Annualized Volatility

Measures how much portfolio returns fluctuate over time.

### Sharpe Ratio

Measures return relative to risk.

```text
(expected return - risk free rate) / volatility
```

### Beta

Measures portfolio sensitivity compared to a benchmark.

### Maximum Drawdown

Measures the largest drop from peak to trough.

### Historical VaR 95%

Estimates a likely maximum one-day loss under historical conditions.

### Concentration Score

Measures how concentrated the portfolio is.

Simple idea:

```text
sum(weight^2)
```

Higher score means lower diversification.

### Correlation Matrix

Shows how assets move relative to each other.

---

## Business Rules

- weights must sum to 1.0 (allow small tolerance like 0.999 to 1.001)
- duplicate tickers are not allowed
- unsupported tickers must return clear errors
- benchmark is required for beta
- stale data should trigger refresh
- analysis can fail gracefully if historical data is insufficient
- each request should be idempotent if retried by the client

---

## Event Design

### If using Kafka

Topics:

- `analysis.requested`
- `marketdata.refresh.requested`
- `analysis.completed`
- `analysis.failed`

Example event:

```json
{
	"analysis_id": "an_123",
	"portfolio_id": "pf_456",
	"period": "1y",
	"benchmark": "SPY",
	"requested_at": "2026-03-25T10:00:00Z"
}
```

Kafka is a strong choice if you want to practice:

- topic-based event design
- consumer groups
- replay
- scaling workers independently

### If using RabbitMQ

Queues:

- `risk-analysis-jobs`
- `market-data-refresh-jobs`
- `notifications`

RabbitMQ is a strong choice if you want:

- easier local setup
- classic job queue model
- direct worker consumption
- simpler retry / dead-letter patterns for task processing

### Recommendation

If your main goal is **backend learning with less operational overhead**, start with **RabbitMQ**.

If your main goal is **distributed systems + event-driven architecture**, start with **Kafka**.

---

## Kubernetes Design

You said you want to use K8s, so this project fits well.

### Suggested workloads

#### Deployments

- api-service
- risk-worker
- market-data-service
- notification-worker

#### Stateful/managed dependencies

- PostgreSQL
- Redis
- Kafka or RabbitMQ

### Kubernetes resources

- Deployment
- Service
- ConfigMap
- Secret
- HorizontalPodAutoscaler
- Ingress
- CronJob _(for data refresh jobs if needed)_

### Example namespace

```yaml
namespace: investment-risk
```

### Good K8s learning goals

- separate deployment per component
- health probes
- resource limits
- rolling updates
- autoscaling risk workers
- secrets for API keys
- environment-specific config

---

## Suggested Folder Structure

```text
investment-risk-engine/
├── cmd/
│   ├── api/
│   ├── worker/
│   └── marketdata/
├── internal/
│   ├── analysis/
│   │   ├── domain/
│   │   ├── service/
│   │   ├── repository/
│   │   └── handler/
│   ├── portfolio/
│   ├── marketdata/
│   ├── messaging/
│   ├── cache/
│   ├── config/
│   ├── db/
│   └── common/
├── pkg/
│   └── metrics/
├── deployments/
│   ├── docker/
│   ├── k8s/
│   └── helm/
├── scripts/
├── test/
├── docs/
├── go.mod
└── README.md
```

---

## Suggested Internal Architecture in Go

Use **clean architecture / hexagonal style** lightly, without overengineering.

Example layers:

- `handler`: HTTP handlers
- `service`: use cases and orchestration
- `domain`: entities and business rules
- `repository`: DB access
- `messaging`: queue producer/consumer
- `infra`: database, cache, external API clients

This keeps the core risk logic testable.

---

## Example API

### Create analysis

```http
POST /api/v1/analyses
Content-Type: application/json
```

Body:

```json
{
	"portfolioName": "Growth Portfolio",
	"assets": [
		{ "ticker": "AAPL", "weight": 0.4 },
		{ "ticker": "MSFT", "weight": 0.3 },
		{ "ticker": "SPY", "weight": 0.3 }
	],
	"period": "1y",
	"benchmark": "SPY"
}
```

Response:

```json
{
	"analysisId": "an_123",
	"status": "processing"
}
```

### Get analysis result

```http
GET /api/v1/analyses/an_123
```

Response:

```json
{
	"analysisId": "an_123",
	"status": "completed",
	"result": {
		"annualizedVolatility": 0.18,
		"sharpeRatio": 1.05,
		"beta": 0.92,
		"maxDrawdown": -0.14,
		"historicalVaR95": -0.021,
		"concentrationScore": 0.34
	}
}
```

---

## Processing Flow

1. Client sends portfolio to API.
2. API validates weights and tickers.
3. API creates `analysis_requests` record with status `processing`.
4. API publishes message to Kafka topic or RabbitMQ queue.
5. Worker consumes message.
6. Worker loads or refreshes market data.
7. Worker aligns time-series and calculates returns.
8. Worker calculates metrics.
9. Worker stores `analysis_results`.
10. API returns report when requested.

---

## Handling Failures

A strong project also handles bad scenarios well.

### Examples

- external market API unavailable
- ticker not found
- queue publish failure
- partial market data
- analysis timeout
- duplicate client retries

### Good patterns

- retry with backoff
- dead-letter queue
- idempotency key
- circuit breaker around external APIs
- structured logs
- request correlation IDs

---

## Caching Strategy

Use Redis for:

- recent portfolio analyses
- historical price lookup cache
- request deduplication
- rate-limit helper
- temporary job status

Example:

- key: `analysis:{hash_of_request}`
- TTL: `1h`

If same portfolio + same period + same benchmark are requested again, return cached analysis.

---

## Observability

To make the project feel more senior, add:

- structured logging
- Prometheus metrics
- tracing with OpenTelemetry
- Grafana dashboards

Interesting metrics:

- analysis job duration
- failed analysis count
- market data fetch latency
- queue depth
- cache hit rate

---

## Security

Even for a portfolio project, basic security matters.

- validate payloads strictly
- rate-limit public endpoints
- store API secrets in Kubernetes Secrets
- sanitize logs
- require auth for user-specific portfolios
- use signed JWT if you add authentication

---

## Local Development

A practical local environment:

- API service
- worker
- PostgreSQL
- Redis
- Kafka or RabbitMQ

Run with Docker Compose first.

### Example services

- `api`
- `worker`
- `postgres`
- `redis`
- `rabbitmq` or `kafka`

---

## Deployment Strategy

### Local

Use Docker Compose

### Staging / Production-like

Use Kubernetes

Good progression:

1. build locally
2. containerize services
3. deploy to local K8s with Kind or Minikube
4. add Helm or Kustomize
5. configure autoscaling for workers

---

## Suggested Roadmap

### Week 1

- project setup in Go
- define entities
- create REST API skeleton
- connect PostgreSQL
- create portfolio and analysis tables

### Week 2

- integrate RabbitMQ or Kafka
- implement async job flow
- build worker consumer
- store analysis status

### Week 3

- implement market data ingestion
- implement volatility, Sharpe, max drawdown
- save analysis results

### Week 4

- implement beta, VaR, concentration score
- add Redis cache
- add Docker Compose
- write tests

### Week 5

- add Kubernetes manifests
- deploy API + worker
- configure health checks
- add ConfigMaps and Secrets

### Week 6

- improve observability
- add retry/dead-letter flow
- polish README and architecture docs

---

## Testing Strategy

### Unit tests

- metric calculations
- weight validation
- return series alignment
- concentration score

### Integration tests

- DB repositories
- queue producer/consumer
- market data ingestion

### End-to-end tests

- create analysis request
- consume job
- persist result
- fetch result from API

---

## What will make this project impressive on GitHub

- clean README
- architecture diagram
- sample requests/responses
- Docker Compose for local run
- Kubernetes manifests
- good tests
- clear explanation of tradeoffs between Kafka and RabbitMQ
- realistic failure handling

---

## Suggested MVP Decision

For your first version, I recommend:

- **Go**
- **Gin**
- **PostgreSQL**
- **Redis**
- **RabbitMQ**
- **Docker Compose**
- **Kubernetes manifests**

Why RabbitMQ first:

- easier to implement
- excellent for async job workers
- faster to finish MVP

Then later create a **Kafka branch/version** to show architectural evolution.

---

## Future Improvements

- Monte Carlo simulation
- stress testing scenarios
- sector/factor exposure analysis
- report export to PDF
- user auth and multi-tenant portfolios
- webhook notifications
- benchmark comparison charts
- frontend dashboard in React or Next.js

---

## Resume / Portfolio Description

You can describe this project like this:

> Built a backend investment risk analysis platform in Golang using asynchronous workers, PostgreSQL, Redis, RabbitMQ/Kafka, and Kubernetes. The system ingests historical market data, calculates risk metrics such as volatility, Sharpe ratio, beta, VaR, and max drawdown, and exposes scalable APIs for portfolio analysis.

---

## Nice Repository Names

- `investment-risk-engine`
- `portfolio-risk-platform`
- `risklens`
- `golang-risk-analyzer`
- `quant-risk-backend`

---

## Final Recommendation

If your goal is to learn and also finish a strong project, use this setup:

- **Golang API + Worker**
- **RabbitMQ first**
- **PostgreSQL**
- **Redis**
- **Docker Compose**
- **Kubernetes**
- add **Kafka later** as a second version

That gives you:

- backend depth
- async architecture
- infra experience
- a project that looks real and valuable

---

## License

MIT
