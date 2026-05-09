-- user-service Postgres schema.
-- Mounted into postgres container's /docker-entrypoint-initdb.d/.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ── Enums ─────────────────────────────────────────────────────────────────────
DO $$ BEGIN
  CREATE TYPE user_status AS ENUM ('ACTIVE','SUSPENDED','DELETED');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- vehicle_type is shared with the reservation service.
-- In a split-DB deployment each service owns its copy; here we share one Postgres.
DO $$ BEGIN
  CREATE TYPE vehicle_type AS ENUM ('CAR','MOTORCYCLE');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ── Driver profile ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS user_profile (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  external_user_id  text UNIQUE NOT NULL,       -- sub claim from super-app JWT
  full_name         text NOT NULL,
  phone_e164_enc    bytea,                       -- pgcrypto MSISDN (for SMS)
  email_enc         bytea,                       -- pgcrypto email
  status            user_status NOT NULL DEFAULT 'ACTIVE',
  version           int NOT NULL DEFAULT 1,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_profile_status ON user_profile(status) WHERE status != 'DELETED';

-- ── Vehicle plate registry ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS vehicle (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  driver_id    uuid NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
  nopol        text NOT NULL,                    -- normalised: upper-case, no spaces
  vehicle_type vehicle_type NOT NULL,
  is_default   boolean NOT NULL DEFAULT false,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_vehicle_driver_nopol UNIQUE (driver_id, nopol)
);
CREATE INDEX IF NOT EXISTS idx_vehicle_driver ON vehicle(driver_id);

-- ── gRPC idempotency key store ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS idempotency_key (
  scope             text NOT NULL,
  key               text NOT NULL,
  response_payload  bytea,
  status_code       int,
  created_at        timestamptz NOT NULL DEFAULT now(),
  expires_at        timestamptz NOT NULL,
  PRIMARY KEY (scope, key)
);
CREATE INDEX IF NOT EXISTS idx_idem_expires ON idempotency_key(expires_at);

COMMIT;
