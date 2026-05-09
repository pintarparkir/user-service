# Feature 04 — Profile cache

**Status:** ✅ shipped
**Owner:** user-service

## Scope

Redis read-through cache fronting `GetUserByID` and `GetUserByExternalID`.
Reduces Postgres pressure for the hot path: every authenticated mini-app request
loads the driver profile via `UpsertDriver` → cached lookup.

**In:**
- Cache key: `user:profile:<id>`.
- TTL: 5 minutes.
- Cache-aside pattern (read-then-set on miss).
- Invalidate on `UpdateUser` and `UpsertDriver` write success.

**Out:**
- Vehicle list caching (low cardinality, low frequency — direct DB read is fine).
- Negative caching (avoid masking real "user deleted" responses).

## Why 5 minutes

- Most mini-app sessions are < 5 min. Cache lifetime ≈ session lifetime.
- Drivers rarely update their own profile — invalidation hits are rare.
- Stale-read window (post-update, pre-invalidate-replication) is bounded.

## Tasks

- [x] `pkg/redis` connection pool + facade
- [x] Cache-aside in `usecase.GetUserByID`
- [x] Invalidate on update / upsert
- [x] Graceful fallback when Redis unreachable (warn-and-continue, not error)

## Acceptance criteria

- Second `GetUserByID` within 5 min hits Redis (verifiable via `redis-cli MONITOR`).
- After `UpdateUser`, the next `GetUserByID` returns the *new* value (no stale cache).
- Killing Redis mid-request still allows `GetUserByID` to succeed (DB fallback).
