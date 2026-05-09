# Feature 02 — Vehicle registry

**Status:** ✅ shipped
**Owner:** user-service
**Tracking:** `ROADMAP.md → user-service → MVP → Vehicle registry`

## Scope

A driver registers one or more license plates (nopol) under their profile. The
reservation-service later reads the plate list when the driver creates a booking.

**In:**
- `RegisterVehicle(driver_id, nopol, vehicle_type, is_default)` — idempotent on `(driver_id, nopol)`.
- `ListVehicles(driver_id)` — returns plates ordered by `is_default DESC, created_at ASC`.
- Nopol normalisation: uppercase, strip spaces.
- Format validation: `^[A-Z]{1,2}[0-9]{1,4}[A-Z]{0,3}$`.

**Out:**
- Plate uniqueness across drivers (intentional — one car may be registered by multiple drivers, e.g. fleet vehicles).
- VIN / chassis number capture.
- Photo upload of plate.

## API contract

### gRPC (s2s)
```proto
rpc RegisterVehicle(RegisterVehicleRequest) returns (Vehicle);
rpc ListVehicles(ListVehiclesRequest) returns (ListVehiclesResponse);

enum VehicleType { VEHICLE_TYPE_UNSPECIFIED=0; CAR=1; MOTORCYCLE=2; }

message Vehicle {
  string id = 1;
  string driver_id = 2;
  string nopol = 3;
  VehicleType vehicle_type = 4;
  bool is_default = 5;
  google.protobuf.Timestamp created_at = 6;
}
```

### REST (mini app)
```
POST /v1/me/vehicles
Body: { "nopol": "B1234ABC", "vehicle_type": "CAR", "is_default": true }
→ 201 { id, nopol, vehicle_type, is_default, created_at }

GET /v1/me/vehicles
→ 200 { "vehicles": [...] }
```

## Data model

```sql
CREATE TABLE vehicle (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  driver_id    uuid NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
  nopol        text NOT NULL,
  vehicle_type vehicle_type NOT NULL,
  is_default   boolean NOT NULL DEFAULT false,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_vehicle_driver_nopol UNIQUE (driver_id, nopol)
);
```

Idempotent INSERT:
```sql
INSERT INTO vehicle (driver_id, nopol, vehicle_type, is_default)
VALUES ($1, $2, $3, $4)
ON CONFLICT (driver_id, nopol) DO UPDATE SET
  vehicle_type = EXCLUDED.vehicle_type,
  is_default   = EXCLUDED.is_default
RETURNING ...;
```

## Tasks

- [x] `model.Vehicle` + `VehicleType` enum
- [x] `model.RegisterVehicleRequest.Validate` (regex on normalised nopol)
- [x] `repository.VehicleRepository` interface
- [x] postgres impl with ON CONFLICT upsert
- [x] `usecase.RegisterVehicle` (verify driver exists → upsert)
- [x] `usecase.ListVehicles` (returns empty slice, not error, for new drivers)
- [x] gRPC handlers + proto mappers
- [x] REST handlers `POST /v1/me/vehicles`, `GET /v1/me/vehicles`
- [x] Unit tests + integration test

## Acceptance criteria

- POSTing the same nopol twice with different `is_default` flips the flag, returns
  the *same* `id`, and creates no duplicate row.
- POSTing `"B 1234 ABC"` and `"b1234abc"` both store `"B1234ABC"`.
- Invalid nopol (`"123"`, `"FOO"`, `""`) → 400 with `VALIDATION` error code.
- Listing for a brand-new driver returns `{ "vehicles": [] }`, not 404.
- Default vehicle (if any) is first in the returned list.
