# Presentation Guide — ParkirPintar Backend Solution

**Format expected:** ~15 min walkthrough + Q&A.
**Audience:** Senior engineers + tech leads at Telkomsel reviewing the assessment.

This file is the **presenter's notes** — the story to tell, the slides to show, the questions to anticipate.

---

## 0. Pre-presentation setup (5 min before)

- Open Miro board: https://miro.com/app/board/uXjVHZEcZYM=/
- Open VS Code at `parkirpintar/` repo
- Have terminal ready with `make up` already running (avoid the 30s startup wait)
- Have `README.md` open in browser (mermaid + SVG diagrams render properly)
- Have one tab on the GitHub repo

---

## 1. The 15-minute walkthrough

### Minute 0–2 — Frame the problem

> "The soal is **lite, simple, fast** parking reservation for one building. Not Uber-for-parking, not multi-tenant marketplace. So I optimised for: cheap MVP, fast time-to-market, runnable end-to-end via `docker compose up`. Defer multi-building, defer service mesh, defer Kafka. Single-region cost target $30–40/month."

Show: Miro board zoomed to **HLD diagram**.

### Minute 2–5 — Architecture in one breath

> "Six microservices behind a Gin gateway. Internal calls are gRPC over HTTP/2 (required by soal). Async events via RabbitMQ — that's the Telkomsel B.3 standard. PostgreSQL is the source of truth, Redis is for distributed locks + read cache. OpenTelemetry traces everything."

Point at:
- Gateway → 6 services (with cluster boundaries)
- Workers (outbox publisher + no-show expirer) running independently
- External boundaries: Midtrans, FCM, Cloud Secret Manager

> "Why 6 not 7? I merged `search` and `presence` because they both query inventory state — same cache, same DB read pool. Saves ~$8/mo and one gRPC hop. Justified in ADR-003."

### Minute 5–8 — The hardest problem: double-booking

This is the make-or-break technical conversation.

> "The risk: two drivers click 'reserve' on the same spot at the same millisecond. Naïve solutions like SELECT-then-INSERT have a race window. Service-level locks like Redis are vulnerable to clock drift, network partition, holder crash."

> "My defense in depth: **two layers**."
> 1. **Redis SETNX lock** — fail fast under contention, 30s TTL with Lua compare-and-delete release. Redis is *not* the source of truth here; it's an optimization to reduce DB write contention.
> 2. **PostgreSQL EXCLUDE constraint** — this is the authoritative defense. `EXCLUDE USING gist (spot_id WITH =, hold_window WITH &&)` makes overlapping reservations *impossible at the storage layer*. Even if Redis is split-brained, PG rejects the conflicting INSERT with `pgerrcode 23P01`."

Show: `data/init.sql` line 70-75 + `internal/reservation/repository/postgres/insert.go`.

> "ADR-004 documents this trade-off."

### Minute 8–11 — Idempotency + transactional outbox

> "Soal mandates idempotency for CreateReservation and Checkout. I implemented it as a gRPC interceptor (`pkg/grpcserver/interceptors.go`) backed by a Postgres `idempotency_key` table scoped per (RPC method, key). Replays return the cached response with zero side effects."

> "For event publishing — say `reservation.confirmed` to RabbitMQ — I solved the dual-write problem with the **transactional outbox pattern**. Domain code inserts both the row and an outbox event in the same DB transaction. A background poller drains unsent rows in batches with `FOR UPDATE SKIP LOCKED` and publishes to RabbitMQ, then marks `published_at`. Guarantees at-least-once delivery. Consumers must be idempotent — and they are, because we route by aggregate ID."

Show: Miro **Reservation Sequence diagram** — point to step 6-7 (INSERT reservation + outbox in same tx) and the bottom dashed flow (async drain).

### Minute 11–13 — Resilience + cost story

> "Failure mode matrix is in README §8.1. Key points:"
> - Redis down → fall back to PG advisory lock. Still safe via EXCLUDE.
> - Notification down → log warn, continue. It's non-core. Outbox retries later.
> - Payment gateway down → reservation completes, invoice stays OPEN, settle async via reconciler.
> - Postgres down → hard 503. It's the system of record; no fallback possible.

> "Cost: ~$30–40/month MVP on Cloud Run + Cloud SQL `db-f1-micro` + Upstash Redis free + self-host RabbitMQ on `e2-micro`. Scales to 10× by going `db-g1-small` + read replica. To 100×: switch RabbitMQ → Kafka, partition reservations by spot_id hash."

### Minute 13–14 — Show the code briefly

Open one file: `internal/reservation/usecase/create_reservation.go`.

> "This is the critical path — 80 lines. Reads top-down: idempotency replay check, transaction open, assign spot, acquire Redis lock, insert reservation (EXCLUDE catches double-book), append outbox event, commit. Then out-of-tx: open invoice via gRPC, release Redis lock. Each layer's responsibility is clear."

Also open `pkg/pricing/pricing.go`.

> "Pricing engine is pure functional — zero I/O, fully unit-testable. Ten tests cover all soal edge cases: 30-min parking, exact-hour boundary, midnight crossing, cancel within/after grace, no-show, overstay. The engine is plug-pluggable — adding EV surcharge or membership pricing is a new `Rule` impl, no changes to existing rules."

### Minute 14–15 — Close + invite questions

> "What I'd revisit with more time:
> 1. Replace stub Midtrans with real Snap integration (E2E payment test currently skipped).
> 2. Real geofence library (`s2geometry-go`) instead of Haversine.
> 3. Chaos test (pumba) for Redis/RabbitMQ failure injection in CI.
> 4. Load test with k6 to validate the 200 RPS target."

> "Questions?"

---

## 2. Q&A — anticipated questions + crisp answers

### Q: "Why not Kafka?"
> RabbitMQ is the Telkomsel B.3 standard. Kafka is overkill at our scale (<50 events/sec MVP), operationally heavier, ~$50+/mo managed. Migration path is clean — outbox pattern decouples publisher from broker, only the publisher impl changes. ADR-001 documents this.

### Q: "Why merge search and presence?"
> Both are read-heavy, query the same inventory state, cache from the same Redis. No state-machine overlap. Merging saves one Cloud Run min instance (~$8/mo) and one gRPC hop on the common availability call. Trigger to split: when either grows distinct ML/personalization logic. ADR-003.

### Q: "What if two drivers race for the same spot in user-selected mode?"
> Three layers: (1) frontend `HoldSpot` Redis SETNX with 10s TTL during selection, (2) standard reservation flow Redis lock with 30s TTL during commit, (3) PG EXCLUDE constraint as authoritative. Loser gets `gRPC ALREADY_EXISTS` with retry-after.

### Q: "Why no service mesh?"
> Six services don't justify the operational cost of Istio/Linkerd at MVP. mTLS via Envoy SDS at the gateway covers our security need. Re-evaluate at >20 services or when we need policy-based traffic shifting.

### Q: "How do you handle clock skew between Redis and Postgres for the lock TTL?"
> The Redis lock is a *performance optimization*, not a correctness mechanism. Even if the lock expires early (clock skew makes Redis think the holder died) and a second caller acquires it, the PG EXCLUDE constraint catches the overlap. The lock just makes the system fail-fast under contention; correctness comes from PG.

### Q: "How do you do retries safely?"
> Three rules: (1) only retry idempotent RPCs (CreateReservation has `Idempotency-Key`; OpenInvoice keys on reservation_id), (2) exponential backoff with jitter via `cenkalti/backoff/v4`, (3) circuit breaker per dependency via `sony/gobreaker` — open after 5 consecutive failures in 10s, half-open after 30s. See `pkg/circuitbreaker`.

### Q: "What's the SLO?"
> README §10.2: confirm success 99.5% rolling 30d, p95 latency <250ms, payment webhook 99.9%. Burn-rate alerts at 2% in 1h fast-burn for confirm rate; 1% in 15min for payment webhook.

### Q: "How do you encrypt PII?"
> User service's `phone_e164` and `email` are encrypted at rest via `pgp_sym_encrypt` (pgcrypto). Key sourced from Cloud Secret Manager, rotated externally. The repository handles encrypt/decrypt round-trip so usecases work with plaintext. See `internal/user/repository/postgres/`.

### Q: "Why pgcrypto symmetric and not envelope encryption with KMS?"
> Simpler MVP — one key, one tool. Envelope encryption is the next step when handling >100k users or when key rotation cadence is short. Trigger documented in ADR-005 alternatives.

### Q: "Why hand-written mocks instead of mockgen?"
> Mockgen requires a code-gen step in CI which adds ~30s and a dependency. The User service has 6 interface methods — hand-writing the mock is ~50 LOC and reads better. Switch to mockgen when interface count grows past ~20 across the codebase.

### Q: "How did you handle the soal v1.0(1) revision?"
> They removed the cancellation policy bullet entirely. No-show fee dropped from 10,000 to 5,000 IDR (the only no-show figure left in the soal). Cancellation fee structure became my assumption — surfaced in README §2.1 and ADR-005 (revised). Net change: 1 config default, ~14 file alignments. Clean blast radius because tariffs were already env-driven.

### Q: "How do I know the EXCLUDE constraint actually works?"
> Integration test: `test/integration/reservation_billing_test.go::TestIntegration_DoubleBook_Rejected` — inserts two reservations with overlapping `hold_window` on the same spot in the same tx. Second one fails with `pgerrcode 23P01` → repository maps to `apperror.ErrDoubleBook` → handler returns gRPC `ALREADY_EXISTS`. Test asserts the error.

### Q: "If I had to add multi-building support, what's the migration?"
> ADR-002 documents this. Steps: (1) Add `building_id` column with default `'BLDG-001'`, (2) Add `building` table with geofence + timezone + pricing override, (3) Update queries to filter by `building_id`, (4) Update gateway routing `/v1/buildings/:id/availability`. ~2 dev-days.

### Q: "What if Midtrans webhook is replayed by an attacker?"
> The webhook handler verifies HMAC-SHA512 signature against Midtrans's secret before any state mutation. Replays with valid signature still hit the idempotency check (we key on Midtrans's `transaction_id`). Replays with invalid signature get rejected at signature step.

### Q: "How would you change this for 10× scale?"
> Bump Postgres to `db-g1-small`. Add read replica for `presence-search`. Redis primary+replica. RabbitMQ stays single-node up to ~5k events/sec. Cloud Run autoscales horizontally — code is already stateless. No architectural changes needed for 10×.

### Q: "Show me how a wrong-spot scenario flows."
> `internal/reservation/usecase/check_in.go` lines 26-44 — if `req.ActualSpotID != current.SpotID`, we still allow check-in (driver shouldn't get stuck if attendant directs them) but emit a `reservation.wrong_spot.v1` event for ops audit. No monetary penalty per soal — that decision was a deliberate read of "wrong-spot penalty" as a *test-required scenario name*, not a fee specification.

---

## 3. If they want to see the code running

Have these commands ready (paste into terminal):

```bash
# Show health
make smoke

# Reserve a spot
curl -X POST http://localhost:8080/v1/reservations \
  -H "Authorization: Bearer dev-driver-1" \
  -H "Idempotency-Key: $(uuidgen)" \
  -H "Content-Type: application/json" \
  -d '{"vehicle_type":"CAR","mode":"SYSTEM_ASSIGNED"}'

# Try to double-book (should fail)
RESV_ID=$(curl ... | jq -r .id)
curl -X POST http://localhost:8080/v1/reservations \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"vehicle_type":"CAR","mode":"USER_SELECTED","preferred_spot_id":"F1-C-001"}'
# Expect 409 Conflict

# Run pricing tests
go test -v ./pkg/pricing/...
```

---

## 4. Things to NOT do during the presentation

- Don't open IDE files at random — pre-decide which files to show.
- Don't apologise for stubs (payment, presence) — they're intentional MVP boundaries with clear migration paths.
- Don't read the README aloud — they've read it. Tell them the *story* the README backs up.
- Don't get defensive about the cancellation-fee assumption — own it: "soal v1.0(1) removed the bullet, here's my reasoning, here's the ADR, tariffs are env-driven so business can change them."
- Don't promise things you didn't build (e.g. "I have full Midtrans integration") — say "I stubbed Midtrans, here's the contract, here's what real integration looks like."

---

## 5. Body language / framing tips

- **Lead with trade-offs, not features.** "I chose RabbitMQ over Kafka because…" reads more senior than "I used RabbitMQ".
- **Quote the soal back to them.** "The soal said 'lite, simple, fast' — that's why I avoided service mesh." Shows you read carefully.
- **Cite your own docs.** "ADR-004 covers this in more detail" — they'll appreciate the discipline.
- **Time-box yourself.** Aim 14 min so you have buffer; running over is unprofessional.

Good luck!
