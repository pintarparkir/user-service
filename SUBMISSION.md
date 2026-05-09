# Submission Workflow

How to take this codebase from local → submitted to the assessor.

## Pre-flight checklist

Before pushing, verify:

- [ ] `git clean -fd` removes empty leftover folders (`services/`, `proto/`, `deploy/`, etc. from restructuring)
- [ ] `make test-unit` passes locally (no infra required)
- [ ] `make up && make smoke` brings stack up healthy
- [ ] `configs/.env` is **not** committed (`.gitignore` covers this)
- [ ] No real Midtrans / production secrets anywhere — grep for `sk_live`, `prod`, real phone numbers
- [ ] `README.md` §18 has the **Git URL** and **Miro board URL** filled in
- [ ] `data/init.sql` schema version matches what's in the running DB (re-run from clean)

```bash
# One-shot pre-flight
make fmt && make vet && make test-unit && \
  make up && make smoke && make down && \
  echo "✓ ready to push"
```

## Repo creation

GitHub (free, public):

```bash
# 1. Create repo on github.com/new — call it `parkirpintar`, public, no README
# 2. Locally:
cd parkirpintar
git init
git add .
git commit -m "feat: initial ParkirPintar backend solution

Solution Development Assessment 2026 - Smart Parking Marketplace.
Stack: Go 1.22, gRPC over HTTP/2, PostgreSQL 16, Redis, RabbitMQ.
Boilerplate: cmd/internal/pkg layout per Telkomsel B.4 standard.
See README.md for HLD, LLD, ERD, and ADRs."

git branch -M main
git remote add origin git@github.com:<your-username>/parkirpintar.git
git push -u origin main
```

GitLab equivalent — same flow, swap remote URL.

## Recommended commit history (if you want a clean log)

If you want the reviewer to see a story instead of one giant commit, replay roughly:

```bash
git reset --soft <root>     # keep changes, undo single commit

# 1. Scaffolding
git add Makefile Dockerfile go.mod .gitignore .gitlab-ci.yml configs/.env.example
git commit -m "chore: project scaffolding (Makefile, Dockerfile, go.mod, CI)"

# 2. Proto contracts
git add api/proto/
git commit -m "feat(proto): gRPC contracts for reservation, billing, payment, user, common"

# 3. Reusable pkg/
git add pkg/
git commit -m "feat(pkg): reusable libraries (configs, logger, db, redis, otel,
grpcserver/client, pricing engine, locking, idempotency, outbox, rabbitmq,
circuitbreaker, scheduler)"

# 4. Reservation domain
git add internal/reservation/ cmd/reservation/
git commit -m "feat(reservation): full domain implementation with gRPC handler,
sqlx repository, no-show worker, transactional outbox"

# 5. Billing domain
git add internal/billing/ cmd/billing/
git commit -m "feat(billing): full domain with pricing engine integration,
RabbitMQ event consumer, invoice ledger"

# 6. User sample microservice
git add internal/user/ cmd/user/ mock/repository/user_repository_mock.go data/migrations/
git commit -m "feat(user): sample microservice — gRPC + REST dual transport,
pgcrypto PII encryption, optimistic locking"

# 7. Other services (stubs)
git add cmd/{gateway,payment,presence,notification,worker}/
git commit -m "feat(services): gateway BFF (Gin), payment stub (Midtrans QRIS),
presence stub, notification consumer, worker sidecar"

# 8. Infrastructure
git add data/ deployments/ scripts/
git commit -m "chore(deploy): docker-compose stack, k8s manifests,
postgres/redis/rabbitmq configs, OTel collector, gen scripts"

# 9. Tests
git add test/ internal/**/*_test.go pkg/**/*_test.go
git commit -m "test: unit tests (pricing rules, idempotency, validation),
integration (reservation→billing, user CRUD), E2E (happy path, double-book,
cancellation policy)"

# 10. Documentation
git add README.md docs/ SUBMISSION.md PRESENTATION.md
git commit -m "docs: README with HLD/LLD/ERD, 6 ADRs, library decision,
4 architecture diagrams (SVG + Miro), submission + presentation guides"

git push -f origin main   # only safe because we just created the repo
```

> **Time investment:** ~10 min for the replay. Skip if you'd rather submit a single comprehensive commit — assessor will read the README either way.

## Branch protection (optional but professional)

If you have time:

```bash
# Tag the submission point
git tag -a v1.0.0-submission -m "Assessment 2026 cycle 1H submission" && git push --tags

# Mark main as protected on GitHub:
# Settings → Branches → Add rule for `main` → require PR review (1) + status checks
```

This shows you understand release hygiene without adding friction to the assessor.

## What to share with the assessor

In your submission email / form, include:

1. **Git URL** — `https://github.com/<you>/parkirpintar`
2. **Miro board URL** — `https://miro.com/app/board/uXjVHZEcZYM=/`
3. **Quickstart** — one line: "`make up && make smoke && make test-unit`"
4. **Reading order** — start with `README.md` §3 (HLD), then `internal/reservation/usecase/create_reservation.go` for the critical-path code
5. **Presentation ready** — point to `PRESENTATION.md` for 15-min walkthrough + Q&A prep

## Post-submission hygiene

- Don't push more commits to `main` after submission — confuses the assessor.
- If you find a bug, push to `hotfix/post-submission` branch and email the assessor a heads-up.
- Keep the Miro board read-only after submission (Settings → Sharing → "Can view" for "Anyone with link").
