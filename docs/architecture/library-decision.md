# Library Decision — Mapping to Telkomsel B.3 Standard

This document records how user-service's dependency choices map back to the internal
**B.3 Library** standard. Anything not on the standard list has explicit justification.

## Direct mapping (B.3 → user-service)

| B.3 Category | B.3 Recommendation | user-service uses | Notes |
|---|---|---|---|
| Web | Gin 1.9.1 | `gin-gonic/gin` 1.10.0 | Mini-app REST surface (`/v1/me*`) |
| Logging | Zap + Lumberjack | `uber-go/zap` 1.27 + `natefinch/lumberjack` 2.2 | `pkg/logger` |
| Configuration | Godotenv | `joho/godotenv` 1.5.1 + `caarlos0/env` 3.5 | `pkg/configs` |
| Monitoring | Prometheus | `prometheus/client_golang` 1.19 | RED metrics + business KPIs |
| RDBMS | SQLx | `jmoiron/sqlx` 1.4 + `lib/pq` 1.10 | All repositories use `pkg/db/postgres` |
| Key-value | Go Redis | `go-redis/redis/v8` 8.11.5 | `pkg/redis` (profile cache) |
| Queue | RabbitMQ | `rabbitmq/amqp091-go` 1.10 | Event publishing (when emitting) |
| Scheduler | GoCron | `go-co-op/gocron/v2` 2.2 | Background jobs (idempotency-key TTL sweep) |
| CLI | Cobra | `spf13/cobra` 1.8 | Reserved for admin tools (seed, migrate) |
| Unit Test | Testify | `stretchr/testify` 1.9 | All `_test.go` files |

## Additions outside B.3 (with justification)

| Library | Why | LOC saved vs hand-rolling |
|---|---|---|
| `google.golang.org/grpc` 1.64 | Required by soal: "service-to-service via gRPC over HTTP/2". Not on B.3 because B.3 focuses on REST/HTTP. | N/A — required |
| `go.opentelemetry.io/otel` 1.27 | Modern tracing standard. B.3 has Prometheus for metrics; OTel covers distributed traces. | ~500 LOC of hand-rolled propagation |
| `sony/gobreaker` 1.0 | Per-dependency circuit breaker on outbound gRPC. | ~150 LOC of state-machine + tests |
| `cenkalti/backoff/v4` 4.3 | Exponential backoff with jitter, context-aware. | ~80 LOC + correctness tests |
| `golang-migrate/migrate/v4` 4.17 | Versioned schema migration. | ~300 LOC of homegrown |
| `google/uuid` 1.6 | RFC 4122 v4 IDs. | n/a — std library doesn't expose UUID |
| `caarlos0/env` 3.5 | Tag-based env binding; pairs with godotenv (which only loads, doesn't decode). | ~120 LOC of reflection |
| `uptrace/opentelemetry-go-extra/otelsqlx` 0.3 | OTel auto-instrumentation for sqlx queries. | ~80 LOC of manual span wrapping |
| `golang/mock` 1.6 | Mock generator for repository interfaces (testing). | n/a — code generation |

## JWT verification — stdlib, no library

The mini app calls user-service over HTTPS with an RS256 JWT issued by the super-app.
We verify the signature using only `crypto/rsa`, `crypto/sha256`, and `encoding/pem`
from the standard library (see `pkg/jwt/jwt.go`, ~80 LOC).

Why not `golang-jwt/jwt`:
- We only consume one token type (RS256, super-app's). No need for HS256/ES256/EdDSA.
- We never *issue* tokens. Identity is delegated.
- One fewer dependency on the public-facing path; smaller attack surface.

If the requirements grow (e.g., issuing service-to-service JWTs, accepting multiple
algorithms), revisit and adopt `golang-jwt/jwt/v5`.

## Frameworks: Why Go Native (per B.2)

B.2 benchmarks show Go Native wins on latency (1.6 ms p95 vs 3.1 ms Beego, 6.6 ms GoFr)
and binary size (11 M vs 18 M vs 43 M).

For user-service: pure Go Native + `google.golang.org/grpc` for the gRPC server,
`gin-gonic/gin` for the REST surface to the mini app. The latency cost of Gin
(~1.5 ms over net/http) is acceptable for one network hop.

## Packages we deliberately did **not** add

| Package | Why not |
|---|---|
| ORM (GORM, ent) | Hides query plans. sqlx + raw SQL is clearer for `pgcrypto` and partial indexes. |
| Kafka | Overkill at our scale; B.3 standard is RabbitMQ. |
| Vault library | Cloud Secret Manager is preferred at the deploy target. |
| Service mesh (Istio/Linkerd) | Operationally heavy for 4 services. Re-evaluate at >20 services. |
| `golang-jwt/jwt` | One token type, stdlib is enough (see "JWT verification" above). |
