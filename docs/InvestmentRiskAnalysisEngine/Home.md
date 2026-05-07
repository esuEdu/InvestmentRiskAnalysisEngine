# Investment Risk Analysis Engine

A backend-focused **portfolio risk analysis platform** built with **Golang**, designed to analyze investment portfolios using historical market data and asynchronous processing.

> This vault replaces the README. All project documentation lives here.

---

## System

- [[Architecture]] — System design, components, and data flow
- [[Database Schema]] — Tables and field definitions
- [[Risk Metrics]] — Financial metrics calculated by the engine
- [[Infrastructure]] — Docker, Kubernetes, NGINX setup
- [[Development Guide]] — Local setup and workflow
- [[Project Plan]] — Roadmap, current status, future improvements

---

## Services

### API Service (`cmd/api` · `:8080`)
- [[API Reference]] — Endpoints and request/response contracts
- [[Services/API Service/Plan|Plan]] — Roadmap and future improvements

### Worker Service (`cmd/worker` · RabbitMQ consumer)
- [[Worker Service]] — Implementation reference
- [[Services/Worker Service/Implementation Plan|Implementation Plan]] — Step-by-step build guide (Phase 3)
- [[Services/Worker Service/Plan|Plan]] — Roadmap and future improvements

### Market Data Service (`cmd/marketdata` · `:8081`)
- [[Market Data Service]] — Service overview and endpoints
- [[Services/Market Data Service/Plan|Plan]] — Roadmap and future improvements

### Portfolio Service (`cmd/portfolio` · `:8082` · *planned*)
- [[Services/Portfolio Service/Plan|Plan]] — Design and future roadmap

---

## Project Goal

Build a backend system that receives a **portfolio of assets** and calculates risk metrics asynchronously, emphasising:

- Backend architecture and distributed systems
- Asynchronous job processing
- Cloud-native infrastructure (Kubernetes)

### Tech Stack

| Layer | Technology |
|---|---|
| Language | Golang |
| Web framework | Gin |
| Database | PostgreSQL (via pgx + sqlc) |
| Messaging | RabbitMQ |
| Config | Viper |
| Logging | Uber Zap |
| Container | Docker |
| Orchestration | Kubernetes + NGINX Ingress |
