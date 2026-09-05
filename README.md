# url-shortener

A Go URL shortener with async click analytics, built as a one-day, hands-on learning project. Redis serves double duty as both the primary data store and the async job queue, kept intentionally simple in scope (no external broker like RabbitMQ/Kafka — that's a deliberate follow-up, not part of this build).

## Architecture

```
                 ┌─────────────┐
  POST /shorten  │             │  SET url:<code> -> URL
 ───────────────►│   Go (Gin)  │─────────────────────────►  ┌───────┐
                 │   Server    │                             │ Redis │
  GET /{code}    │             │  GET url:<code>            │       │
 ───────────────►│             │─────────────────────────►  └───┬───┘
                 │             │  302 redirect                   │
                 │             │  LPUSH clicks (async)           │
                 └─────────────┘                                 │
                       │                                          │
                       │ /metrics                          BRPOP  │
                       ▼                                          ▼
                 ┌───────────┐                           ┌────────────────┐
                 │Prometheus │◄──────────────────────────│ Worker goroutine│
                 └─────┬─────┘        scrapes app          (drains clicks, │
                       │                                    INCR counters) │
                       ▼                                   └────────────────┘
                 ┌───────────┐
                 │  Grafana  │  (dashboards, provisioned)
                 └───────────┘
```

The redirect path never blocks on click analytics — a click event is pushed onto a Redis list (`LPUSH`) and a background worker goroutine drains it (`BRPOP`) and increments per-code counters asynchronously.

## Features

- `POST /shorten` — generates a short code, stores the code→URL mapping in Redis (`SETNX`, with TTL support)
- `GET /{code}` — looks up the code, issues a 302 redirect, and asynchronously records the click
- Background worker with graceful shutdown (context + channels) draining the click queue
- Prometheus metrics at `/metrics`:
  - `urlshortener_shortens_total`, `urlshortener_redirects_total` — counters
  - `urlshortener_click_queue_depth` — gauge (queue backpressure signal)
  - `urlshortener_http_request_duration_seconds` — histogram, labeled by `method`/`path`/`status`
  - `urlshortener_redis_operation_duration_seconds` — histogram, labeled by `operation` (`get`/`set`/`lpush`/`brpop`), with tightened sub-millisecond buckets tuned for local Redis latency
- Grafana dashboard (fully provisioned — zero manual clicking required): redirect rate, shorten rate, queue depth, and p95 latency panels
- GitHub Actions CI: `go build`, `go vet`, `go test` on every push/PR to `main`

## Tech stack

Go, Gin, Redis, Prometheus, Grafana, Docker Compose, GitHub Actions.

## Getting started

```bash
docker compose up
```

This brings up the full stack: the app, Redis, Prometheus, and Grafana.

| Service     | URL                          |
|-------------|-------------------------------|
| App         | http://localhost:8080         |
| Prometheus  | http://localhost:9090         |
| Grafana     | http://localhost:3000 (admin/admin) |

Grafana's Prometheus datasource and the dashboard are provisioned automatically — no manual setup needed after `docker compose up`.

### Example usage

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
# -> {"code": "abc123", "short_url": "http://localhost:8080/abc123"}

curl -L http://localhost:8080/abc123
# -> 302 redirect to https://example.com
```

## Configuration

| Env var      | Description                  | Default (compose) |
|--------------|-------------------------------|--------------------|
| `REDIS_ADDR` | Redis host:port               | `redis:6379`       |
| `PORT`       | App listen port                | `8080`             |

## Project structure

```
.
├── .github/workflows/ci.yml          # build, vet, test on push/PR
├── grafana/provisioning/
│   ├── datasources/datasource.yml    # Prometheus datasource (pinned UID)
│   └── dashboards/
│       ├── dashboard.yml             # provider config
│       └── urlshortener.json         # exported dashboard
├── internal/
│   ├── handler/                      # Gin HTTP handlers
│   ├── metrics/                      # Prometheus metric declarations
│   └── repository/                   # Redis access layer (timed, instrumented)
├── prometheus.yml                    # scrape config
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Observability notes

- Redirect/shorten rate panels use `rate()` over a 5m window to smooth bursty traffic given a 15s scrape interval.
- p95 latency blends all endpoints together (`sum by (le)`); breaking it out per-endpoint (`sum by (le, path)`) is a natural next step if per-route visibility becomes necessary.
- Known limitation: the `brpop` Redis-operation timing includes time spent blocked waiting for a queue item, not just Redis round-trip time — it's a proxy for "how long clicks sit before being drained," not pure Redis latency.
- Redis error paths (real connection/timeout errors) are timed and recorded in the same histogram as successes; only `redis.Nil` (key-miss / empty-queue, a normal outcome) is excluded from timing.

## Known limitations / next steps

- No unit or integration tests yet.
- No code TTL/expiry policy implemented.
- No in-memory LRU cache for hot codes.
- No load testing performed yet (`hey`/`k6` planned).
- Redis is both store and queue by design — a real broker (RabbitMQ/Kafka) and gRPC are deliberate out-of-scope follow-ups.
