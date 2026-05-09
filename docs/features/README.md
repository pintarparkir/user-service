# Features — user-service

One markdown file per feature. Each spec carries:

1. **Status** — shipped / in-progress / planned (mirrored in `../../../ROADMAP.md`)
2. **Scope** — what's in, what's out
3. **API contract** — REST or gRPC signatures
4. **Data model** — tables / columns touched
5. **Tasks** — checklist used during implementation
6. **Acceptance criteria** — how we know it's done
7. **Open questions** — TBD items

## Index

| File                              | Status | Summary                                      |
|-----------------------------------|--------|----------------------------------------------|
| `01-driver-lazy-registration.md`  | ✅      | Auto-create profile on first JWT-bearing call|
| `02-vehicle-registry.md`          | ✅      | Register / list vehicles, idempotent on plate|
| `03-pii-encryption.md`            | ✅      | pgcrypto on phone & email; key in Secret Mgr |
| `04-profile-cache.md`             | ✅      | Redis read-through, 5-min TTL                |
| `05-jwt-middleware.md`            | ✅      | Stdlib RS256 verifier on `/v1/me*`           |
| `06-idempotency.md`               | ✅      | Postgres-backed replay cache for write RPCs  |
| `10-soft-delete-reactivate.md`    | ⏳      | Status flow `ACTIVE ↔ SUSPENDED ↔ DELETED`   |
| `11-pii-key-rotation.md`          | 📋      | Dual-read decrypt, re-encrypt CLI tool       |
