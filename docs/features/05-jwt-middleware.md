# Feature 05 — JWT middleware

**Status:** ✅ shipped
**Owner:** user-service

## Scope

Mini-app HTTP requests carry a super-app-issued RS256 JWT. user-service verifies
the signature and extracts identity claims before routing to the handler.

**In:**
- Stdlib-only RS256 verifier (`pkg/jwt`, ~80 LOC).
- Extracts `sub` (external_user_id) and `phone` (E.164) from the JWT payload.
- `exp` claim enforced (rejects expired tokens with 401).
- Dev mode: empty `SUPER_APP_JWT_PUBLIC_KEY_PEM` skips signature verification (payload still parsed).
- Returns 401 with body `{ "error": "UNAUTHENTICATED", "message": "<reason>" }` on failure.

**Out:**
- Issuing tokens (super-app does this).
- HS256 / ES256 / EdDSA support.
- JWK rotation / multiple keys (single PEM in env for now).

## Why stdlib only

We consume exactly one token type. Pulling `golang-jwt/jwt` for that adds an
external dep on the public-facing path. Stdlib `crypto/rsa` + `encoding/pem`
already do PKCS#1 v1.5 SHA-256 verification cleanly. See
`docs/architecture/library-decision.md` for full rationale.

## Tasks

- [x] `pkg/jwt.Parse(token, pubKeyPEM) -> *Claims, error`
- [x] `internal/user/handler/http/middleware.go.jwtMiddleware`
- [x] `cmd/user/main.go` reads `cfg.SuperAppJWTPubKey` and passes to `RegisterUserHandler`
- [x] `.env.example` documents `SUPER_APP_JWT_PUBLIC_KEY_PEM`
- [ ] Replace `pkg/jwt` with library if requirements grow (multiple algos, JWK rotation)

## Acceptance criteria

- Request with no `Authorization` header → 401.
- Request with `Authorization: Bearer <expired>` → 401, message mentions expired.
- Request with valid token → handler runs; `c.GetString("driver_id")` is non-empty.
- `SUPER_APP_JWT_PUBLIC_KEY_PEM` empty → signature is *not* checked but `exp` still is.
