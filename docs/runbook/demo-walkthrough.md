# Demo Walkthrough — user-service

A copy-paste runbook to demo user-service end-to-end during the assessment review.
Each step is one curl command + the expected response shape.

## Setup (~30 s)

```bash
# 1. Bring up shared infra (postgres, redis, rabbitmq, otel)
cd ../infra
docker compose up -d

# 2. Run user-service against it
cd ../user-service
cp configs/.env.example configs/.env
make migrate-up                       # apply schema to user_service DB
make run                              # starts gRPC :9094 + REST :8080
```

Health check:
```bash
curl -s http://localhost:8080/healthz
# → {"status":"ok"}
```

## Scenario 1 — Lazy driver registration on first hit (60 s)

The mini app POSTs with the super-app JWT. user-service decodes the JWT, extracts
`sub` and `phone`, and `UpsertDriver`s — creating the row on first contact.

```bash
# Generate a dev JWT (signature verification is skipped when SUPER_APP_JWT_PUBLIC_KEY_PEM is empty)
PAYLOAD=$(echo -n '{"sub":"super-user-001","phone":"+628123456789","exp":9999999999}' | base64)
TOKEN="eyJhbGciOiJSUzI1NiJ9.${PAYLOAD}.devsig"

# First call: row is created behind the scenes
curl -s http://localhost:8080/v1/me -H "Authorization: Bearer $TOKEN" | jq .
# → {"id":"<uuid>","external_user_id":"super-user-001","phone_e164":"+628123456789","status":"ACTIVE","version":1,...}

# Second call: same row, no insert (idempotent)
curl -s http://localhost:8080/v1/me -H "Authorization: Bearer $TOKEN" | jq .id
```

**Talking points:**
- Lazy registration: no signup screen — driver is created on first authenticated call.
- Keyed on `external_user_id` (the JWT `sub`), not on phone. Phone can change.
- Driver row survives across logins; `UpsertDriver` just returns the existing row.

## Scenario 2 — Vehicle registration (45 s)

```bash
# Register a car
curl -s -X POST http://localhost:8080/v1/me/vehicles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nopol":"B 1234 ABC","vehicle_type":"CAR","is_default":true}' | jq .
# → {"id":"<uuid>","nopol":"B1234ABC","vehicle_type":"CAR","is_default":true,...}

# Re-register the same plate — idempotent, returns same row (toggle is_default)
curl -s -X POST http://localhost:8080/v1/me/vehicles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nopol":"b1234abc","vehicle_type":"CAR","is_default":false}' | jq .

# List my vehicles
curl -s http://localhost:8080/v1/me/vehicles -H "Authorization: Bearer $TOKEN" | jq .
```

**Talking points:**
- Nopol is normalised before storage: `"B 1234 ABC"` and `"b1234abc"` become `"B1234ABC"`.
- Validation regex enforces Indonesian plate format `^[A-Z]{1,2}[0-9]{1,4}[A-Z]{0,3}$`.
- `(driver_id, nopol)` UNIQUE → re-registration becomes an UPDATE, not a duplicate row.

## Scenario 3 — gRPC s2s call (notification-service simulation, 30 s)

notification-service calls `GetUserById` over gRPC to resolve MSISDN before sending SMS.

```bash
# Use grpcurl against the gRPC port
DRIVER_ID=$(curl -s http://localhost:8080/v1/me -H "Authorization: Bearer $TOKEN" | jq -r .id)

grpcurl -plaintext \
  -d "{\"id\":\"$DRIVER_ID\"}" \
  localhost:9094 parkirpintar.user.v1.UserService/GetUserById | jq .
# → {id, externalUserId, fullName, phoneE164:"+628123456789", ...}
```

**Talking points:**
- s2s calls go through `pkg/grpcserver` which has the OTel + idempotency interceptors.
- Phone is decrypted on read via `pgcrypto` `pgp_sym_decrypt(...)`.
- The same `pkg/jwt` middleware does NOT run on gRPC — s2s callers are trusted
  (network-level isolation in production).

## Scenario 4 — Idempotency replay on UpdateUser (30 s)

```bash
IDEM=$(uuidgen)

# First call — flips full_name
curl -s -X PUT http://localhost:8080/v1/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $IDEM" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Budi Santoso","expected_version":1}' | jq .

# Same call again with same Idempotency-Key — returns cached response, no second update
curl -s -X PUT http://localhost:8080/v1/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Idempotency-Key: $IDEM" \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Budi Santoso","expected_version":1}' | jq .
```

**Talking points:**
- Backed by `idempotency_key` Postgres table, scoped per `(scope=method, key)`.
- 24-hour TTL, swept by background cleanup.
- gRPC interceptor in `pkg/grpcserver/interceptors.go` does the replay check.

## Scenario 5 — PII encryption in flight (30 s, optional)

```bash
docker exec -it parkir-postgres psql -U postgres -d user_service \
  -c "SELECT id, external_user_id, phone_e164_enc FROM user_profile LIMIT 3;"
# phone_e164_enc shows as bytea ciphertext, not plaintext +62...
```

**Talking points:**
- pgcrypto `pgp_sym_encrypt(phone, $PG_CRYPTO_KEY)` on write.
- Decrypted only at the repository read boundary.
- Key rotation strategy: dual-read (old + new key) for one release window, then
  re-encrypt + drop old key from Secret Manager. Documented in `README.md §5`.

## Cleanup

```bash
make down
cd ../infra && docker compose down -v
```

## Total time: ~3–4 min for all scenarios
