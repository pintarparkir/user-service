# High-Level Design — ParkirPintar

This document describes the system architecture, communication patterns, and deployment
topology for the ParkirPintar smart-parking marketplace.

## 1. System Context

ParkirPintar runs as a mini-app inside a super-app shell (Tencent Cloud Mini Program Platform).
User identity and auth are delegated to the host super-app via RS256 JWT. Each service
verifies the JWT independently — there is no BFF or API gateway layer.

```
┌────────────────────────────────────────────────────────────────────┐
│                     Super-App (Mobile)                             │
│                                                                    │
│  Other Mini-Apps  │   ParkirPintar Mini-App   │  Wallet etc.       │
└─────────────────────────────┬──────────────────────────────────────┘
                              │ HTTPS + Bearer JWT (RS256)
                              │ wx.request
                              │
           ┌──────────────────┼──────────────────────┐
           │                  │                      │
           ▼                  ▼                      ▼
  ┌─────────────────┐ ┌─────────────────┐  ┌────────────────────┐
  │  user-service   │ │reservation-svc  │  │ billing/payment-svc│
  │  Cloud Run      │ │  Cloud Run      │  │   Cloud Run        │
  │  REST HTTP :8080│ │  REST HTTP :8080│  │   REST HTTP :8080  │
  └────────┬────────┘ └────────┬────────┘  └────────────────────┘
           │ gRPC (s2s)        │ gRPC (s2s)
           │◀──────────────────┤
           │                   │── gRPC ──▶ billing-service
           │                   │
           │◀── gRPC ──────────┘ (MSISDN lookup)
           │
  ┌────────▼────────────────────────────────────────────────────────┐
  │                    Shared Infrastructure                        │
  │  PostgreSQL (Cloud SQL)  │  Redis (Memorystore)  │  RabbitMQ   │
  └─────────────────────────────────────────────────────────────────┘
                                          │
                               ┌──────────▼───────────┐
                               │  notification-service │
                               │  Cloud Run (consumer) │
                               │  RabbitMQ + gRPC      │
                               └──────────────────────┘
```

## 2. Communication Patterns

### 2a. Mini App → Services (REST HTTP)

The mini app calls each service directly over HTTPS. Every request carries:
- `Authorization: Bearer <jwt>` — RS256 token issued by super-app
- `Idempotency-Key: <uuid>` — required on mutating requests

Each service runs a JWT middleware that:
1. Parses and verifies the RS256 signature using the super-app's public key PEM
2. Extracts `sub` (external_user_id) and `phone` (E.164) claims
3. Calls `UpsertDriver` (user-service only) or passes `X-Driver-Id` downstream

### 2b. Service-to-Service (gRPC over HTTP/2)

| Caller               | Callee           | RPC                |
|----------------------|------------------|--------------------|
| reservation-service  | billing-service  | OpenInvoice        |
| reservation-service  | user-service     | GetUserById        |
| notification-service | user-service     | GetUserById        |

All s2s gRPC calls:
- Propagate `Idempotency-Key` as gRPC metadata
- Use OTLP trace context propagation
- Circuit-breaker via `gobreaker` on caller side

### 2c. Async Events (RabbitMQ)

Exchange: `parkirpintar.events` (topic, durable)

| Routing Key                  | Producer             | Consumer(s)              |
|------------------------------|----------------------|--------------------------|
| `reservation.confirmed.v1`   | reservation-service  | notification-service     |
| `reservation.cancelled.v1`   | reservation-service  | notification-service     |
| `reservation.expired.v1`     | reservation-service  | notification-service     |
| `billing.invoice.closed.v1`  | billing-service      | notification-service     |
| `payment.paid.v1`            | payment-service      | reservation, notification|

**Delivery guarantee:** Outbox pattern in PostgreSQL. A background goroutine polls
`outbox_event` rows, publishes to RabbitMQ, then marks `published_at`. Guarantees
at-least-once delivery; consumers must be idempotent.

## 3. Reservation Sequence — Happy Path

```
Driver        user-service    reservation-svc    Redis     Postgres    billing-svc   RabbitMQ
  │                │                │              │            │            │            │
  │── POST ────────────────────────▶│              │            │            │            │
  │  /v1/reservations               │              │            │            │            │
  │                │                │── SETNX ────▶│            │            │            │
  │                │                │  lock:spot   │            │            │            │
  │                │                │◀── OK ───────│            │            │            │
  │                │                │── BEGIN ──────────────────▶            │            │
  │                │                │── INSERT reservation ──────▶            │            │
  │                │                │── INSERT outbox_event ─────▶            │            │
  │                │                │── COMMIT ──────────────────▶            │            │
  │                │                │                            │            │            │
  │                │                │── gRPC OpenInvoice() ───────────────────▶            │
  │                │                │◀─────────── invoice ───────────────────│            │
  │                │                │── DEL lock:spot ────▶│                 │            │
  │◀── 201 ────────────────────────│               │            │            │            │
  │                │                │                            │            │            │
  │                │         ┌── outbox poller goroutine ──────────────────────────────────
  │                │         │  SELECT unsent FROM outbox_event                           │
  │                │         │── publish reservation.confirmed.v1 ────────────────────────▶
  │                │         │  UPDATE outbox_event SET published_at=now()                │
  │                │         └──                                                          │
  │                │                │             notification-service consumes event     │
  │                │◀── gRPC GetUserById ──────────────────────────────────────────────── │
  │                │──── MSISDN ────────────────────────────────────────────────────────── │
  │                │                │             SMS sent to driver                      │
```

## 4. JWT Auth Flow (per-service, no gateway)

```
Mini Program
  │  Authorization: Bearer <super-app-jwt>
  ▼
user-service HTTP middleware
  │  1. Parse JWT header/payload (base64url decode)
  │  2. Verify RS256 sig with SUPER_APP_JWT_PUBLIC_KEY_PEM
  │  3. Check exp claim
  │  4. Extract sub=external_user_id, phone=E.164
  ▼
UpsertDriver middleware (user-service only)
  │  UpsertDriver(phone, external_user_id) → driver_id (idempotent)
  │  Sets driver_id in Gin context
  ▼
Handler: uses driver_id from context
```

Other services (reservation, billing) verify the JWT the same way but read
`X-Driver-Id` passed by the mini app after it receives it from user-service on first
registration, or they call user-service gRPC to resolve it.

## 5. User Service REST API (mini app)

Base path: `/v1`

| Method | Path              | Description                      |
|--------|-------------------|----------------------------------|
| GET    | /v1/me            | Get own driver profile           |
| PUT    | /v1/me            | Update own profile (name, email) |
| GET    | /v1/me/vehicles   | List own registered vehicles     |
| POST   | /v1/me/vehicles   | Register a new vehicle           |

All endpoints require `Authorization: Bearer <jwt>`.

gRPC server (`:9094`) remains for s2s callers (notification-service, reservation-service).

## 6. Deployment Topology

### Dev (Docker Compose)
Single-node, all services + infra on developer laptop. Each service builds from its own
`Dockerfile`. Shared `docker-compose.yml` at repo root orchestrates postgres, redis,
rabbitmq, otel-collector.

### Production (Google Cloud Run)

| Service              | Cloud Run | Min Instances | Notes                            |
|----------------------|-----------|---------------|----------------------------------|
| user-service         | Yes       | 0             | scale-to-zero OK (fast startup)  |
| reservation-service  | Yes       | 1             | avoid cold-start on critical path|
| billing-service      | Yes       | 0             |                                  |
| notification-service | Yes       | 0             | event-driven, latency tolerant   |
| payment-service      | Yes       | 0             |                                  |

Cloud Run specifics:
- Port: read from `PORT` env var (Cloud Run injects this; default 8080 locally)
- gRPC: Cloud Run supports HTTP/2 natively — no special config needed
- Secrets: `SUPER_APP_JWT_PUBLIC_KEY_PEM`, `PG_CRYPTO_KEY` via Secret Manager

Infrastructure:
- **PostgreSQL**: Cloud SQL `db-f1-micro`, 7-day PITR, private IP via Cloud SQL Proxy sidecar
- **Redis**: Memorystore Basic 1 GB
- **RabbitMQ**: CloudAMQP Lemur (free tier) for MVP, or a single `e2-micro` GCE VM with persistent disk

### CI/CD
GitHub Actions: lint (`golangci-lint`) → unit tests → integration tests (testcontainers)
→ `docker build` → push to Artifact Registry → `gcloud run deploy`.

Branch protection: PRs require green CI. `main` auto-deploys to staging; manual approval gate to prod.
