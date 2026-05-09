# Low-Level Design — user-service

Detail level: package layout, key types, transactions, error mapping.
Other services have their own LLDs in their respective `docs/architecture/` folders.

## 1. Package Layout

```
user-service/
├── api/proto/user/v1/         ← protobuf (UserService gRPC contract)
├── cmd/user/main.go           ← service entry point (gRPC + REST)
├── configs/                   ← .env.example
├── data/                      ← init.sql + golang-migrate migrations
├── internal/user/
│   ├── model/                 ← domain types, request DTOs, validators (no I/O)
│   ├── repository/            ← repository interface + postgres impl
│   ├── usecase/               ← business logic (orchestrates repo + cache)
│   └── handler/
│       ├── grpc/              ← UserService gRPC adapters (s2s callers)
│       └── http/              ← REST adapters for mini app (/v1/me*)
├── pkg/                       ← cross-cutting libs (logger, otel, redis, jwt, sms, …)
├── mock/                      ← gomock-generated test doubles
├── test/integration/          ← testcontainers-based integration suite
└── deployments/               ← service-specific docker-compose + k8s manifests
```

The `infra/` folder at the repo root brings up shared dev infrastructure (postgres,
redis, rabbitmq, otel). user-service connects to those over `host.docker.internal`.

## 2. Domain Types

```go
// internal/user/model/user.go
type UserStatus string
const (
    UserStatusActive    UserStatus = "ACTIVE"
    UserStatusSuspended UserStatus = "SUSPENDED"
    UserStatusDeleted   UserStatus = "DELETED"
)

type User struct {
    ID             string
    ExternalUserID string     // sub claim from super-app JWT
    FullName       string
    PhoneE164      string     // E.164, decrypted from phone_e164_enc
    Email          string     // decrypted from email_enc
    Status         UserStatus
    Version        int        // optimistic lock
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// internal/user/model/vehicle.go
type VehicleType string
const (
    VehicleTypeCar        VehicleType = "CAR"
    VehicleTypeMotorcycle VehicleType = "MOTORCYCLE"
)

type Vehicle struct {
    ID          string
    DriverID    string
    Nopol       string       // normalised: upper-case, no spaces
    VehicleType VehicleType
    IsDefault   bool
    CreatedAt   time.Time
}
```

## 3. Critical Path — `UpsertDriver` (lazy registration)

Called by every authenticated request via the HTTP middleware (`upsertDriverMiddleware`)
and by gRPC clients (notification-service) on each event.

```go
// internal/user/usecase/upsert_driver.go
func (u *userUsecase) UpsertDriver(ctx context.Context, req model.UpsertDriverRequest) (*model.User, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }
    return u.repo.GetOrCreateByMSISDN(ctx, req.PhoneE164, req.ExternalUserID, req.FullName)
}
```

```go
// internal/user/repository/postgres/upsert.go (sketch)
// 1. SELECT by external_user_id  → return on hit (hot path, common case)
// 2. INSERT user_profile         → on success, return.
// 3. On UNIQUE-violation race:
//    SELECT by external_user_id again → return that row (peer just inserted).
```

Why the read-first → insert → re-read pattern:
- The hot path (subsequent calls) is a single SELECT — fast.
- `INSERT ... ON CONFLICT DO NOTHING RETURNING` would silently lose the row when
  another goroutine inserted first. The explicit re-read is unambiguous.
- The whole sequence is wrapped in a single short transaction for consistent timing.

## 4. Critical Path — `RegisterVehicle`

```go
// usecase: validate, normalise nopol, verify driver exists, ON CONFLICT update
INSERT INTO vehicle (driver_id, nopol, vehicle_type, is_default, ...)
VALUES (...)
ON CONFLICT (driver_id, nopol) DO UPDATE SET
  vehicle_type = EXCLUDED.vehicle_type,
  is_default   = EXCLUDED.is_default
RETURNING ...
```

Why upsert and not insert-only:
- The driver may correct their `vehicle_type` (CAR ↔ MOTORCYCLE) for a typoed plate.
- `is_default` toggles by re-registering with the same plate.
- Idempotent at the protocol level: the same nopol always maps to the same row.

## 5. Caching — profile read-through

`pkg/redis` acts as a 5-minute-TTL cache fronting `GetUserByID` and `GetUserByExternalID`.
Cache key: `user:profile:<id>`. Invalidation: `UpdateUser` and `UpsertDriver` `DEL` the
key after a successful repo write so the next read repopulates from Postgres.

The cache is best-effort. If Redis is unreachable, the usecase falls through to
Postgres without erroring (a `Warn` is logged from `cmd/user/main.go` on initial ping
failure; per-request misses are silent).

## 6. JWT Middleware (HTTP)

`internal/user/handler/http/middleware.go` runs before every `/v1/me*` handler:

1. Extract `Authorization: Bearer <token>` (401 if missing / malformed).
2. Decode + verify with `pkg/jwt.Parse(token, cfg.SuperAppJWTPubKey)`.
   - If `pubKeyPEM == ""` (local dev) signature check is skipped — payload is still parsed.
3. `c.Set("external_user_id", claims.Sub); c.Set("phone_e164", claims.Phone)`.
4. `upsertDriverMiddleware` immediately calls `UpsertDriver` and stores `driver_id`.
5. Handlers read `driver_id` via `c.GetString(ctxDriverID)`.

This means `/v1/me` is always operating on a known driver row — no special-case
"first request" code path in handlers.

## 7. Idempotency (gRPC server interceptor)

```go
// pkg/grpcserver — UnaryInterceptor
func UnaryInterceptor(store Store, methods []string) grpc.UnaryServerInterceptor {
    return func(ctx, req, info, handler) (any, error) {
        if !slices.Contains(methods, info.FullMethod) {
            return handler(ctx, req)
        }
        key := KeyFromMD(ctx)              // X-Idempotency-Key from gRPC metadata
        if cached, ok := store.Get(ctx, info.FullMethod, key); ok {
            return cached, nil
        }
        resp, err := handler(ctx, req)
        if err == nil {
            store.Put(ctx, info.FullMethod, key, resp, 24*time.Hour)
        }
        return resp, err
    }
}
```

Idempotent methods are listed explicitly in `cmd/user/main.go`:
- `CreateUser`, `UpdateUser`, `UpsertDriver`, `RegisterVehicle`.

`GetUserByID` and `ListVehicles` are read-only and skip the interceptor.

## 8. Error Mapping (domain → gRPC / HTTP)

| Source                                | gRPC code            | HTTP status | Notes                              |
|---------------------------------------|----------------------|-------------|------------------------------------|
| `apperror.AppError{Code:"VALIDATION"}`| `INVALID_ARGUMENT`   | 400         | Field-level validation             |
| `apperror.AppError{Code:"NOT_FOUND"}` | `NOT_FOUND`          | 404         | Driver / vehicle not found         |
| `apperror.AppError{Code:"CONFLICT"}`  | `ALREADY_EXISTS`     | 409         | Optimistic-lock version mismatch   |
| `apperror.AppError{Code:"UNAUTHENTICATED"}` | `UNAUTHENTICATED` | 401     | JWT missing / invalid              |
| `pgerrcode.UniqueViolation` (idem)    | (return cached)      | (cached)    | Idempotency replay                 |
| `context.DeadlineExceeded`            | `DEADLINE_EXCEEDED`  | 504         | —                                  |
| Anything else                         | `INTERNAL`           | 500         | Logged with stack at WARN          |

## 9. Config (`pkg/configs`)

`caarlos0/env` tag-based binding with `joho/godotenv` loading `configs/.env` at boot.
Per-environment `.env` is git-ignored; `.env.example` is committed and serves as the
schema-of-record. See section §10 of `README.md` for the full env-var table.

## 10. Logging & Tracing

- **Logging**: `uber-go/zap` JSON output to stdout; trace_id + span_id injected from
  the OTel context (see `pkg/logger`).
- **Tracing**: OTLP exporter via `pkg/otel`; gRPC client/server interceptors come
  from `otelgrpc`; HTTP from `otelgin`.
- **Sampling**: 10 % normal traffic, 100 % of errors (tail-based — done at the
  collector, not in-process).
