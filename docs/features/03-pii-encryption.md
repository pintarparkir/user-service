# Feature 03 — PII encryption at rest

**Status:** ✅ shipped
**Owner:** user-service
**Tracking:** `ROADMAP.md → user-service → MVP → PII encryption`

## Scope

Driver MSISDN and email are encrypted at rest using PostgreSQL `pgcrypto` symmetric
encryption. Plaintext never lands in disk-level pages or logical backups.

**In:**
- `phone_e164_enc bytea` — encrypted MSISDN.
- `email_enc bytea` — encrypted email.
- Encrypt on write at the repository layer using `pgp_sym_encrypt(value, $key)`.
- Decrypt on read using `pgp_sym_decrypt(col, $key)::text`.
- Key sourced from `PG_CRYPTO_KEY` env var → Cloud Secret Manager in prod.

**Out:**
- Field-level encryption of `full_name` (treated as low-risk in this MVP).
- Searchable encryption / hash-based lookup of phone (we look up by `external_user_id`).
- Key rotation tooling (see Feature 11).

## Why pgcrypto and not application-side AES

| Approach | Pros | Cons |
|----------|------|------|
| pgcrypto (chosen) | Encrypted within Postgres, plain SQL queries; key never traverses to non-DB clients | Tied to Postgres |
| App-side AES-GCM | Portable across DBs; key only in app process | Every reader needs key + IV mgmt |

We accept Postgres lock-in. user-service is the only writer of these columns.

## Tasks

- [x] `pgcrypto` extension in `data/init.sql`
- [x] `phone_e164_enc` + `email_enc` columns in `user_profile`
- [x] Repository write path uses `pgp_sym_encrypt(...)`
- [x] Repository read path uses `pgp_sym_decrypt(...)::text`
- [x] `PG_CRYPTO_KEY` config field; `.env.example` warns about rotation
- [ ] Key rotation procedure documented (see Feature 11)

## Acceptance criteria

- `SELECT phone_e164_enc FROM user_profile LIMIT 1;` returns binary, not `+62…`.
- A repository read returns the original plaintext MSISDN.
- Logical backup (`pg_dump`) of the table contains only ciphertext for these cols.
- Rotating `PG_CRYPTO_KEY` without first re-encrypting causes decrypt errors
  (proves the key is actually in use).

## Open questions

- *Are we OK losing query-by-phone capability?* Yes — driver lookup is by
  `external_user_id`. SMS delivery starts from a known driver, so phone is
  decrypt-on-demand, never search-by.
