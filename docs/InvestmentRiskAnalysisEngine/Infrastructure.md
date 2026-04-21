## Local Development (Docker Compose)

`docker-compose.yml` spins up all dependencies locally.

```bash
docker compose up
```

Services:

| Service | Port | Notes |
|---|---|---|
| `api` | 8080 | Go API server |
| `postgres` | 5432 | PostgreSQL database |
| `redis` | 6379 | Cache layer |
| `rabbitmq` | 5672 / 15672 | Message broker + management UI |

---

## Kubernetes

Designed to run in a Kubernetes cluster.

### Workloads

| Kind | Name | Notes |
|---|---|---|
| Deployment | `api-service` | Stateless, horizontally scalable |
| Deployment | `risk-worker` | Scales based on queue depth |
| Deployment | `market-data-service` | Scheduled / on-demand data fetcher |
| StatefulSet | `postgres` | Persistent storage |
| StatefulSet | `rabbitmq` | Durable message queue |
| Deployment | `redis` | Cache / ephemeral state |

### Services

- `api-service` → ClusterIP (exposed via Ingress)
- `rabbitmq`, `postgres`, `redis` → ClusterIP (internal only)

---

## NGINX Ingress

Chosen over a full API gateway for simplicity and Kubernetes-native integration.

NGINX handles:
- External traffic routing to `api-service`
- TLS termination
- Load balancing across API pods
- Path-based routing
- Optional rate limiting

Example Ingress manifest:
```yaml
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

> Kong or another API gateway could replace NGINX later if auth, rate-limiting, or plugin ecosystems are needed at scale.

---

## Observability *(planned)*

| Tool | Purpose |
|---|---|
| Prometheus | Metrics scraping |
| Grafana | Dashboards |
| OpenTelemetry | Distributed tracing |

---

## Deployment Steps

1. Build Docker images for `api`, `worker`, `market-data`
2. Deploy PostgreSQL and RabbitMQ (StatefulSets)
3. Run DB migrations
4. Deploy `api-service`
5. Deploy `risk-worker`
6. Apply NGINX Ingress resource

---

## Related Notes

- [[Architecture]]
- [[Development Guide]]
