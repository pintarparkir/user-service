# Feature 06 — Idempotency on writes

**Status:** ✅ shipped
**Owner:** user-service

## Scope

Every mutating RPC accepts a client-supplied `Idempotency-Key`. Replays with the
same key return the previous response without re-executing the handler.

**In:**
- gRPC server interceptor (`pkg/grpcserver`) that intercepts only the methods
  declared in `cmd/user/main.go`'s `IdempotentMethods` list.
- Postgres-backed store (`pkg/idempotency`), table `idempotency_key` keyed on
  `(scope, key)` where scope = `info.FullMethod`.
- 24-hour TTL via `expires_at` column; sweeper cron deletes expired rows.
- Returns the cached response payload + status code on hit.

**Idempotent methods (current list):**
- `CreateUser`
- `UpdateUser`
- `UpsertDriver`
- `RegisterVehicle`

Read-only RPCs (`GetUserById`, `ListVehicles`) skip the interceptor.

## API contract

gRPC: client passes `Idempotency-Key: <uuid>` as gRPC metadata.
REST: client passes the same as an HTTP header on `PUT /v1/me` and `POST /v1/me/vehicles`.

## Tasks

- [x] `idempotency_key` table in `data/init.sql`
- [x] `pkg/idempotency.PostgresStore` (Get/Put with TTL)
- [x] gRPC unary interceptor wired in `pkg/grpcserver`
- [x] Method list explicitly enumerated in `cmd/user/main.go`
- [ ] Background sweeper cron job (currently TTL is checked on read; cleanup is manual)

## Acceptance criteria

- POSTing `RegisterVehicle` twice with the same `Idempotency-Key` returns the same
  response and creates exactly one row.
- POSTing `RegisterVehicle` with the same key but a *different* body still returns
  the *original* response (we treat key as authoritative — body diff is a client bug).
- Different RPC methods with the same key do not collide (scope differentiates).
- After 24 h, replaying with the same key creates a fresh row (TTL respected).
