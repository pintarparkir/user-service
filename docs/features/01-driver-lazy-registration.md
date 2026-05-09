# Feature 01 — Driver lazy registration

**Status:** ✅ shipped
**Owner:** user-service
**Tracking:** `ROADMAP.md → user-service → MVP → Lazy driver registration`

## Scope

The mini app does not have a "sign up" screen. The driver's identity comes from
the super-app JWT. On the *first* authenticated request to user-service, the driver
profile is created automatically; on every subsequent request, the existing row is
returned.

**In:**
- Idempotent `UpsertDriver(phone, external_user_id, full_name?)` RPC.
- Middleware on every `/v1/me*` route that runs `UpsertDriver` before the handler.
- Race-safe insert: read → insert → re-read on UNIQUE-violation.

**Out:**
- Manual signup form (super-app handles identity).
- Email/password auth (super-app handles auth).
- KYC fields (driver_license, ID number) — captured later if needed.

## API contract

### gRPC (s2s)
```proto
rpc UpsertDriver(UpsertDriverRequest) returns (User);

message UpsertDriverRequest {
  string phone_e164      = 1; // required, E.164
  string external_user_id = 2; // required, super-app JWT sub
  string full_name        = 3; // optional, updated on later calls
}
```

### REST (mini app, transparent)
There is no explicit endpoint. The HTTP middleware calls `UpsertDriver` for every
`/v1/me*` request and stores `driver_id` in the Gin context.

## Data model

```sql
-- user_profile (see migrations/002_user.up.sql)
external_user_id text UNIQUE NOT NULL  -- anchor for upsert
phone_e164_enc   bytea                  -- pgcrypto, populated from JWT phone claim
status           user_status DEFAULT 'ACTIVE'
version          int DEFAULT 1
```

## Tasks

- [x] `model.UpsertDriverRequest` + `Validate()`
- [x] `repository.UserRepository.GetOrCreateByMSISDN` (postgres impl)
- [x] `usecase.UpsertDriver` orchestration
- [x] gRPC handler `UpsertDriver`
- [x] HTTP middleware `upsertDriverMiddleware`
- [x] Idempotency-interceptor coverage on `UpsertDriver`
- [x] Unit tests (mock repo)
- [x] Integration test against real Postgres (testcontainers)

## Acceptance criteria

- Calling `UpsertDriver` with the same `external_user_id` twice returns the same
  row (same `id`, no new INSERT).
- Concurrent `UpsertDriver` calls (10 goroutines, same `external_user_id`) all
  return the same `id` and produce exactly one row.
- The `phone_e164` column in Postgres is encrypted (`bytea`, not visible plaintext).
- On JWT-authenticated request to `/v1/me`, response includes the freshly-created
  driver's profile (no separate registration step).

## Open questions

- *Should `full_name` from JWT (if present) overwrite an existing row's full_name?*
  Current behaviour: only set on first INSERT, ignored thereafter. Revisit if
  super-app starts emitting updated full_name claims.
