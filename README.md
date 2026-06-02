# user-service

[![Security](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=security_rating&token=66ea93348f4a130c2f9af61ca6938a8d5b9b4f9a)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Reliability](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=reliability_rating&token=66ea93348f4a130c2f9af61ca6938a8d5b9b4f9a)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=sqale_rating&token=66ea93348f4a130c2f9af61ca6938a8d5b9b4f9a)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Duplications](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=duplicated_lines_density&token=66ea93348f4a130c2f9af61ca6938a8d5b9b4f9a)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=pintarparkir_user-service&metric=coverage&token=66ea93348f4a130c2f9af61ca6938a8d5b9b4f9a)](https://sonarcloud.io/summary/new_code?id=pintarparkir_user-service)

**Cloud Run:** `https://user-service-725nddkmwq-as.a.run.app`

## Architecture Overview

![Architecture](docs/PintarParkir.architecture.svg)

## E2E Flow

![Flow Diagram](docs/flow.diagram.svg)

## Sequence Diagrams

![E2E Sequence Diagram](docs/sequence-diagram.png)

---



> **Purpose:** Driver profile management — owns user identity, vehicle registry, and MSISDN source for notifications.  
> **Author:** Farid Triwicaksono · **Last Updated:** 2026-05-21

## Project Overview

**ParkirPintar** is a backend mini-app for smart parking within a super-app. It handles:
- Availability queries (spots per floor, per vehicle type)
- Reservation creation (system-assigned or user-selected spots)
- Reservation state transitions (confirm, cancel, check-in, check-out)
- Geofence validation (GPS-based check-in)
- No-show expiration (automatic after 1 hour hold)
- Event publishing (outbox pattern → RabbitMQ)

Five services: **user** (this service), **reservation**, **billing**, **payment**, **notification**.

## Service Scope

**Owns:**
- Driver profile (identity, MSISDN, email, status)
- Vehicle registry (license plates, vehicle type, default vehicle)
- PII encryption at rest (pgcrypto for phone/email)
- Lazy registration via gateway (UpsertDriver on first contact)
- Profile cache (Redis TTL 5 minutes)

**Does NOT own:**
- Authentication (delegated to super-app JWT)
- Reservation logic (reservation-service owns)
- Billing/invoicing (billing-service owns)
- Payment processing (payment-service owns)

**Key invariants:**
- One driver per MSISDN (unique constraint)
- PII encrypted at rest via pgcrypto
- Optimistic locking via `version` column
- Idempotency via `external_user_id` natural key

## At a Glance

| Aspect | Details |
|--------|---------|
| **REST Port** | 8080 (mini-app profile endpoints) |
| **gRPC Port** | 9094 (s2s — called by notification/reservation) |
| **Database** | PostgreSQL 16 (users, vehicles, idempotency_key) |
| **Cache** | Redis 7 (profile cache, TTL 5min) |
| **Message Queue** | N/A (no events published) |
| **External APIs** | None (called by other services) |

## Tech Stack

- **Language:** Go 1.22
- **Web Framework:** Gin (REST) + gRPC
- **Database:** PostgreSQL 16 + sqlx
- **Cache:** Redis 7 (go-redis v8)
- **Logging:** Zap + Lumberjack
- **Observability:** OpenTelemetry (OTLP/gRPC)
- **Testing:** testify/mock, table-driven tests
- **PII Encryption:** pgcrypto (symmetric key from Secret Manager)

## Architecture

### High-Level Design
See [`../docs/architecture/high-level-design/01-user-service.md`](../docs/architecture/high-level-design/01-user-service.md) for:
- Service responsibilities and boundaries
- Lazy registration flow (gateway → UpsertDriver)
- Profile cache strategy

### Low-Level Design
See [`../docs/architecture/low-level-design/01-user-service-lld.md`](../docs/architecture/low-level-design/01-user-service-lld.md) for:
- Layer cake (model → usecase → repository → handler)
- PII encryption/decryption flow
- Cache-aside pattern

### Entity Relationship Diagram
See [`../docs/architecture/erd/01-user-service.md`](../docs/architecture/erd/01-user-service.md) for:
- Table schema (users, vehicles, idempotency_key)
- Unique constraints (phone_e164, external_user_id, nopol per driver)
- Critical indexes

![ParkirPintar ERD](ERD.jpg)

## API Reference

### REST Endpoints (mini-app, all require `Authorization: Bearer <jwt>`)

| Method | Path | Description | Idempotent |
|--------|------|-------------|-----------|
| GET | `/v1/me` | Get current driver profile | Yes |
| PUT | `/v1/me` | Update profile (name, email) | Yes (via version) |
| GET | `/v1/me/vehicles` | List driver's vehicles | Yes |
| POST | `/v1/me/vehicles` | Register new vehicle | Yes (via nopol unique) |

### gRPC Services (s2s, internal only)

| RPC | Input | Output | Purpose |
|-----|-------|--------|---------|
| CreateUser | CreateUserRequest | User | Admin user creation |
| UpsertDriver | UpsertDriverRequest | User | Gateway lazy registration (idempotent on MSISDN) |
| GetUserById | GetUserByIdRequest | User | Lookup by internal ID |
| GetUserByExternalId | GetUserByExternalIdRequest | User | Lookup by super-app identity |
| UpdateUser | UpdateUserRequest | User | Admin profile update (optimistic lock) |
| DeleteUser | DeleteUserRequest | DeleteUserResponse | Soft delete (status → DELETED) |
| ListUsers | ListUsersRequest | ListUsersResponse | Admin list with pagination |
| RegisterVehicle | RegisterVehicleRequest | Vehicle | Register license plate (idempotent on nopol) |
| ListVehicles | ListVehiclesRequest | ListVehiclesResponse | Get driver's vehicles |

## Sample Environment

```bash
# ── App ─────────────────────────────────────────────────────────────────────
APP_NAME=user-service
APP_ENV=local
APP_PORT=8080        # REST port (mini app)
GRPC_PORT=9094       # gRPC port (s2s)

# ── Postgres ────────────────────────────────────────────────────────────────
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=user_service
DB_MAX_OPEN=25
DB_MAX_IDLE=10

# ── Redis (profile cache) ───────────────────────────────────────────────────
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1
REDIS_APP_CONFIG=user-service

# ── Observability ────────────────────────────────────────────────────────────
OTLP_ENDPOINT=localhost:4317

# ── PII encryption ───────────────────────────────────────────────────────────
PG_CRYPTO_KEY=local-dev-pgcrypto-key-change-me

# ── SMS notification ─────────────────────────────────────────────────────────
SMS_SENDER_ID=ParkirPintar

# ── JWT verification ─────────────────────────────────────────────────────────
SUPER_APP_JWT_PUBLIC_KEY_PEM=
```

See `configs/.env.example` for full reference.

## Getting Started

### Prerequisites
- Docker 24+ & Docker Compose v2
- Go 1.22+ (for local development)
- `buf` CLI (for proto regeneration)

### Local Development

```bash
# 1. Clone and setup
git clone <repo> && cd <repo>
cd user-service
cp configs/.env.example configs/.env

# 2. Start shared infra (see https://github.com/pintarparkir/infra)
cd ../infra && podman compose up -d

# 3. Run migrations
cd ../user-service
make migrate-up

# 4. Run the service
make run

# 5. Verify health
curl http://localhost:8080/healthz
```

## Testing

### Unit Tests (no infra)
```bash
make test-unit
```
Covers: usecase logic, profile validation, vehicle normalization.

### Integration Tests (requires postgres/redis)
```bash
make test-integration
```
Covers: repository layer, PII encryption/decryption, cache-aside pattern.

### All Tests
```bash
make test
```

### Coverage
```bash
go test -coverprofile=cov.out ./...
go tool cover -html=cov.out
```
Target: usecase ≥80%, repository ≥60%.

## Debugging

### Logs
```bash
LOG_LEVEL=debug make run
```
Logs are JSON-formatted with trace_id, span_id, request_id.

### Database
```bash
psql postgresql://postgres:postgres@localhost:5432/user_service

# View schema
\dt

# Check PII encryption
SELECT id, phone_e164_enc, pgp_sym_decrypt(phone_e164_enc::bytea, 'key') FROM users LIMIT 1;
```

### Redis
```bash
redis-cli

# Inspect profile cache
KEYS user:profile:*

# Check TTL
TTL user:profile:<user_id>
```

### gRPC
```bash
# Test gRPC health
grpcurl -plaintext localhost:9094 grpc.health.v1.Health/Check

# Call UpsertDriver
grpcurl -plaintext -d '{"phone_e164":"+628123456789","external_user_id":"super-app-123","full_name":"Test Driver"}' \
  localhost:9094 parkirpintar.user.v1.UserService/UpsertDriver
```

## Operations

### Health Checks
```bash
# REST
curl http://localhost:8080/healthz

# gRPC
grpcurl -plaintext localhost:9094 grpc.health.v1.Health/Check
```

### Migrations
```bash
make migrate-up      # Apply all pending migrations
make migrate-down    # Rollback one migration
```

### Cache Invalidation
Profile cache auto-expires after 5 minutes. Manual invalidation:
```bash
redis-cli DEL user:profile:<user_id>
```

## Security Notes

- **Secrets:** Never commit `.env` files. Use Secret Manager in production.
- **PII:** Phone numbers and emails encrypted at rest via pgcrypto. Encryption key stored in Secret Manager.
- **JWT:** Verified by gateway; service trusts `X-Driver-Id` header from gateway.
- **SQL:** All queries parameterized (sqlx prevents injection).
- **Optimistic locking:** `version` column prevents lost updates on concurrent profile edits.

## Business Flow Logic

### User Service in End-to-End Flow

User-service adalah **identity provider** untuk seluruh sistem ParkirPintar. Service ini tidak memiliki flow bisnis mandiri, tetapi dipanggil oleh service lain untuk:

1. **Lazy Registration** — Gateway memanggil `UpsertDriver` saat driver pertama kali akses mini-app
2. **MSISDN Resolution** — Notification-service memanggil `GetUserById` untuk resolve phone number sebelum kirim SMS

```mermaid
sequenceDiagram
    autonumber
    participant Gateway as Mini-App Gateway
    participant UserSvc as User Service
    participant Redis as Redis Cache
    participant DB as Postgres (pgcrypto)
    
    Note over Gateway,DB: Flow 1: Lazy Registration (First Access)
    
    Gateway->>UserSvc: gRPC UpsertDriver(external_user_id, phone, name)
    activate UserSvc
    
    UserSvc->>DB: BEGIN
    UserSvc->>DB: SELECT * FROM user_profile WHERE external_user_id = ?
    
    alt User exists
        DB-->>UserSvc: Return existing user
        UserSvc->>DB: COMMIT
    else User not exists
        UserSvc->>UserSvc: Encrypt phone with pgcrypto
        UserSvc->>DB: INSERT INTO user_profile (external_user_id, phone_enc, status=ACTIVE)
        UserSvc->>DB: COMMIT
    end
    
    UserSvc->>Redis: SET user:profile:{id} TTL=5min
    UserSvc-->>Gateway: User profile
    deactivate UserSvc
    
    Note over Gateway,DB: Flow 2: MSISDN Resolution (SMS Dispatch)
    
    participant NotifSvc as Notification Service
    
    NotifSvc->>UserSvc: gRPC GetUserById(driver_id)
    activate UserSvc
    
    UserSvc->>Redis: GET user:profile:{driver_id}
    
    alt Cache hit
        Redis-->>UserSvc: Cached profile
        UserSvc-->>NotifSvc: phone_e164
    else Cache miss
        UserSvc->>DB: SELECT phone_enc FROM user_profile WHERE id = ?
        UserSvc->>UserSvc: Decrypt phone with pgcrypto
        UserSvc->>Redis: SET user:profile:{id} TTL=5min
        UserSvc-->>NotifSvc: phone_e164
    end
    
    deactivate UserSvc
```

### Key Responsibilities in Cross-Service Flows

| Flow | Role | Trigger |
|------|------|---------|
| Reservation Create | Validate driver exists | Gateway JWT verification |
| SMS Notification | Provide MSISDN | Notification-service gRPC call |
| Vehicle Registration | Store driver's vehicles | Mini-app REST POST /v1/me/vehicles |

---

## Related Documentation

- **Architecture Overview:** [`../docs/README.md`](../docs/README.md)
- **Shared Infra Docs:** [`infra`](https://github.com/pintarparkir/infra)
- **API Documentation:** [`../docs/api-documentation/00-overview.md`](../docs/api-documentation/00-overview.md)
- **Implementation Backlog:** [`../docs/implementation-todo/00-backlog.md`](../docs/implementation-todo/00-backlog.md)

---

_For questions or issues, refer to the troubleshooting section in the main README or open an issue on the repo._
