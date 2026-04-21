# Investment Risk Analysis Engine

A backend-focused **portfolio risk analysis platform** built with **Golang**, designed to analyze investment portfolios using historical market data and asynchronous processing.

> This vault replaces the README. All project documentation lives here.

---

## Navigation

- [[docs/Architecture]] — System design, components, and data flow
- [[API Reference]] — Endpoints, request/response contracts
- [[Database Schema]] — Tables and field definitions
- [[Risk Metrics]] — Financial metrics calculated by the engine
- [[Infrastructure]] — Docker, Kubernetes, NGINX setup
- [[Development Guide]] — Local setup and workflow
- [[Project Plan]] — Roadmap, current status, future improvements

---

## Project Goal

Build a backend system that receives a **portfolio of assets** and calculates risk metrics asynchronously, emphasising:

- Backend architecture and distributed systems
- Asynchronous job processing
- Cloud-native infrastructure (Kubernetes)

### Tech Stack

| Layer         | Technology                  |
| ------------- | --------------------------- |
| Language      | Golang                      |
| Web framework | Gin                         |
| Database      | PostgreSQL (via pgx + sqlc) |
| Cache         | Redis                       |
| Messaging     | RabbitMQ                    |
| Config        | Viper                       |
| Logging       | Uber Zap                    |
| Container     | Docker                      |
| Orchestration | Kubernetes + NGINX Ingress  |
