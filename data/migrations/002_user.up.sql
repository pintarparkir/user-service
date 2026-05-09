-- 002_user.up — user_profile + vehicle tables.
-- Owned by the user service. The reservation service references driver_id by string
-- (loose coupling); FK enforcement intentionally omitted to allow microservice DBs
-- to be split later without painful migration.

BEGIN;

DO $$ BEGIN
  CREATE TYPE user_status AS ENUM ('ACTIVE','SUSPENDED','DELETED');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS user_profile (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  external_user_id  text UNIQUE NOT NULL,
  full_name         text NOT NULL,
  phone_e164_enc    bytea,                   -- pgcrypto MSISDN (for SMS notification)
  email_enc         bytea,                   -- pgcrypto email
  status            user_status NOT NULL DEFAULT 'ACTIVE',
  version           int NOT NULL DEFAULT 1,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_profile_status ON user_profile(status) WHERE status != 'DELETED';

-- vehicle_type reuses the enum already created in 001_reservation.up.
CREATE TABLE IF NOT EXISTS vehicle (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  driver_id    uuid NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
  nopol        text NOT NULL,
  vehicle_type vehicle_type NOT NULL,
  is_default   boolean NOT NULL DEFAULT false,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_vehicle_driver_nopol UNIQUE (driver_id, nopol)
);
CREATE INDEX IF NOT EXISTS idx_vehicle_driver ON vehicle(driver_id);

COMMIT;
