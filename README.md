# Investment Risk Analysis Engine

A backend-focused **portfolio risk analysis platform** built with **Golang**, designed to analyze investment portfolios using historical market data and asynchronous processing.

This project focuses heavily on **backend engineering**, **distributed systems**, and **cloud-native infrastructure**.

Key technologies used:

- **Golang**
- **PostgreSQL**
- **Redis**
- **RabbitMQ**
- **Docker**
- **Kubernetes**
- **NGINX Ingress Controller**

---

# Project Goal

Create a backend system that receives a **portfolio of assets** and calculates **risk metrics** such as:

- Annualized volatility
- Sharpe ratio
- Beta
- Maximum drawdown
- Value at Risk (VaR 95%)
- Concentration risk
- Correlation matrix

The emphasis of this project is **backend architecture and distributed processing**, not just CRUD APIs.

---

# Architecture Overview

The system uses **asynchronous job processing** to compute financial metrics.

Workflow:

Client → API Service → Queue (RabbitMQ) → Risk Worker → Database → API Response

High‑level architecture:

```
Client
  |
  v
NGINX Ingress
  |
  v
API Service (Go)
  |
  v
RabbitMQ
  |
  v
Risk Worker (Go)
  |
  v
PostgreSQL / Redis
```

---

# Core Components

## 1. API Service (Golang)

Responsibilities:

- Expose REST API
- Validate portfolio input
- Persist analysis requests
- Publish jobs to RabbitMQ
- Serve analysis results

Example endpoints:

```
POST /api/v1/analyses
GET  /api/v1/analyses/{id}
GET  /api/v1/portfolios/{id}/analyses
GET  /health
```

---

## 2. Risk Analysis Worker

Background worker that:

1. Consumes jobs from RabbitMQ
2. Fetches historical market data
3. Calculates risk metrics
4. Saves results

Metrics computed:

- volatility
- Sharpe ratio
- beta
- max drawdown
- historical VaR
- concentration score

---

## 3. Market Data Service

Responsible for:

- fetching historical prices
- normalizing data
- refreshing stale data
- storing OHLCV time‑series

Possible data providers:

- AlphaVantage
- TwelveData
- Polygon
- Yahoo Finance APIs

---

## 4. Messaging Layer (RabbitMQ)

RabbitMQ is used to decouple:

- API requests
- heavy calculations

Queues:

```
risk-analysis-jobs
market-data-refresh-jobs
notifications
```

Advantages:

- simple job worker model
- reliable retries
- dead letter queues
- easy local development

---

# Infrastructure

## Kubernetes

The system is designed to run in Kubernetes.

Recommended workloads:

Deployment:

- api-service
- risk-worker
- market-data-service

Services:

- api-service
- rabbitmq
- postgres
- redis

Ingress:

- **NGINX Ingress Controller**

---

# Why NGINX Ingress

This project uses **NGINX Ingress Controller** to expose services externally.

NGINX handles:

- routing traffic to services
- TLS termination
- load balancing
- path routing
- rate limiting (optional)

Example flow:

```
Internet
   |
   v
NGINX Ingress
   |
   v
Go API Service
```

Reasons for choosing NGINX:

- industry standard Kubernetes ingress
- simple configuration
- lightweight compared to full API gateways
- perfect for exposing internal services

API gateways like **Kong** could be added later if the platform evolves.

---

# Suggested Stack

Backend:

- Golang
- Gin or Fiber
- PostgreSQL
- Redis
- RabbitMQ

Infrastructure:

- Docker
- Kubernetes
- NGINX Ingress Controller

Observability (future):

- Prometheus
- Grafana
- OpenTelemetry

---

# Database Design

## portfolios

- id
- user_id
- name
- created_at

## portfolio_assets

- id
- portfolio_id
- ticker
- weight

## analysis_requests

- id
- portfolio_id
- benchmark
- period
- status
- created_at

## analysis_results

- analysis_request_id
- annualized_volatility
- sharpe_ratio
- beta
- max_drawdown
- var_95
- concentration_score
- raw_metrics_json

## historical_prices

- ticker
- price_date
- open
- high
- low
- close
- volume

---

# API Example

Create analysis:

```
POST /api/v1/analyses
```

Body:

```json
{
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

---

# Risk Metrics Implemented

## Volatility

Measures how much portfolio returns fluctuate.

## Sharpe Ratio

Measures return relative to risk.

```
(expected return − risk free rate) / volatility
```

## Beta

Sensitivity to benchmark market movements.

## Maximum Drawdown

Largest drop from peak portfolio value.

## Historical VaR

Estimated worst loss within a confidence level.

## Concentration Score

Measures diversification:

```
sum(weight^2)
```

Higher values mean lower diversification.

---

# Suggested Folder Structure

```
investment-risk-engine
│
├── cmd
│   ├── api
│   ├── worker
│   └── marketdata
│
├── internal
│   ├── analysis
│   ├── portfolio
│   ├── messaging
│   ├── marketdata
│   ├── cache
│   ├── config
│   └── db
│
├── deployments
│   ├── docker
│   └── k8s
│
├── scripts
├── docs
└── README.md
```

---

# Local Development

Run services using Docker Compose:

Services:

- api
- worker
- postgres
- redis
- rabbitmq

Example:

```
docker compose up
```

---

# Kubernetes Deployment

Steps:

1. Build Docker images
2. Deploy database and messaging
3. Deploy API service
4. Deploy workers
5. Configure NGINX ingress

Example ingress:

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: risk-api
spec:
  ingressClassName: nginx
  rules:
  - host: risk.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 8080
```

---

# Future Improvements

Possible next steps:

- Monte Carlo simulation
- portfolio stress testing
- sector exposure analysis
- user authentication
- PDF risk reports
- webhook notifications
- frontend dashboard
- Kafka version for event streaming

---

# Resume Description

Example description:

Built a **Golang backend investment risk analysis platform** using asynchronous workers, RabbitMQ, PostgreSQL, Redis, and Kubernetes with NGINX ingress. The system ingests historical market data and calculates portfolio risk metrics such as volatility, Sharpe ratio, beta, VaR, and maximum drawdown.

---

# License

MIT
