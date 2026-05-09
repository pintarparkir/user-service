# User Service — sample microservice

A complete, runnable sample showing how to build a new domain in this codebase.
Use it as a template when adding new services (driver-vehicle, building, partner, ...).

## What it does
- CRUD for super-app users (driver profile records).
- PII (phone, email) **encrypted at rest** via pgcrypto symmetric keys.
- **Idempotent CreateUser** keyed on `external_user_id` — safe to retry.
- **Optimistic locking** on UpdateUser via `version`.
- **Soft delete** — preserves audit trail.
- **Cache-first reads** via Redis (5-min TTL, invalidated on write).
- **Two transports**: gRPC for internal callers, REST/JSON for ops dashboards.

## Layered architecture (Clean Architecture)

```
internal/user/
├── model/
│   ├── const.go        ← UserStatus, idempotency scopes, pagination caps
│   ├── user.go         ← aggregate (PII-aware fields)
│   ├── request.go      ← CreateUserRequest, UpdateUserRequest, ListUsersRequest
│   ├── response.go     ← ListUsersResponse
│   └── validator.go    ← cheap pre-DB validation (E.164, email format)
│
├── repository/
│   ├── type.go                          ← UserRepository interface
│   └── postgres/
│       ├── type.go                      ← adapter struct + decrypt projection
│       ├── create_user.go               ← INSERT + pgp_sym_encrypt PII
│       ├── get_user_by_id.go            ← SELECT + pgp_sym_decrypt
│       ├── update_user.go               ← partial patch + optimistic lock
│       ├── soft_delete_user.go          ← status='DELETED' (idempotent)
│       └── list_users.go                ← pagination + status filter + total
│
├── usecase/
│   ├── type.go                ← UserUsecase interface
│   ├── init.go                ← constructor (deps wiring)
│   ├── create_user.go         ← idempotent on external_user_id
│   ├── get_user_by_id.go      ← cache-first read
│   ├── update_user.go         ← invalidates cache after success
│   ├── delete_user.go         ← invalidates cache
│   ├── list_users.go          ← paginated
│   ├── create_user_test.go    ← table-driven validation + idempotency
│   └── update_user_test.go    ← optimistic-lock conflict scenario
│
└── handler/
    ├── grpc/
    │   ├── type.go · init.go · mapper.go
    │   ├── create_user.go · get_user.go · update_user.go
    │   ├── delete_user.go · list_users.go
    └── http/
        ├── type.go · dto.go · error_mapper.go
        ├── create_user.go · get_user.go · update_user.go
        ├── delete_user.go · list_users.go
```

## API surface

### gRPC (`api/proto/user/v1/user.proto`)
| RPC | Idempotent | Notes |
|---|---|---|
| `CreateUser` | ✅ on `external_user_id` | Returns existing record on retry |
| `GetUserById` | ✅ (read) | Cache-first |
| `GetUserByExternalId` | ✅ (read) | SSO sign-in path |
| `UpdateUser` | ✅ on `expected_version` | Optimistic lock |
| `DeleteUser` | ✅ | Soft delete; safe to retry |
| `ListUsers` | ✅ (read) | Pagination + status filter |

### REST (`/users`)
| Method | Path | Purpose |
|---|---|---|
| `POST`   | `/users` | Create user (idempotent on `external_user_id`) |
| `GET`    | `/users/:id` | Get by id |
| `GET`    | `/users/by-ext/:ext` | Get by external id |
| `PUT`    | `/users/:id` | Partial update + version check |
| `DELETE` | `/users/:id` | Soft delete |
| `GET`    | `/users?limit=&offset=&status=` | Paginated list |

## Local quickstart
```bash
cp configs/.env.example configs/.env
make up                                          # brings up postgres, redis, all services

# gRPC ping
grpcurl -plaintext localhost:9094 list

# REST
curl -X POST http://localhost:8084/users \
  -H 'Content-Type: application/json' \
  -d '{"external_user_id":"ext-001","full_name":"Farid","phone_e164":"+628111111111","email":"f@example.com"}'

curl http://localhost:8084/users/by-ext/ext-001
```

## Tests
```bash
go test -short ./internal/user/...                                 # unit (mock-backed)
go test -tags=integration ./test/integration/user_test.go          # integration (real PG + pgcrypto)
```

## Why this is a useful sample

- Shows **both gRPC + HTTP** transports sharing one usecase — the only correct pattern when a service has both internal and ops consumers.
- Demonstrates **pgcrypto round-trip** in a clean way: the domain object holds plaintext, the SQL handles encrypt/decrypt.
- Shows **idempotency without an Idempotency-Key header** — leveraging a natural unique key (`external_user_id`) is often simpler than the generic header machinery.
- Wires **all the standard pkg/** components: configs, logger, otel, redis, idempotency interceptor, gRPC server.

## Notes & trade-offs

| Decision | Why | When to revisit |
|---|---|---|
| Single Postgres DB shared with reservation/billing | Simpler MVP ops; no cross-DB tx needed | Split when this service grows separate scaling needs (>500 RPS) |
| pgcrypto symmetric (vs envelope encryption) | Easier MVP; key in env | Move to envelope (KMS) when handling >100k users |
| In-process REST + gRPC in one binary | Simpler deploy; same usecase, no drift | Split when transports need independent scaling/SLO |
| Mocks hand-written | No mockgen dependency in CI | Replace with `mockgen` when interface count grows |
