package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/internal/user/repository"
)

type vehicleRepository struct {
	db *sqlx.DB
}

// NewVehicleRepository wires the Postgres adapter for the VehicleRepository port.
func NewVehicleRepository(db *sqlx.DB) repository.VehicleRepository {
	return &vehicleRepository{db: db}
}

// Register inserts a vehicle plate for a driver.
// ON CONFLICT (driver_id, nopol) updates vehicle_type so re-registration is idempotent.
func (r *vehicleRepository) Register(ctx context.Context, v model.Vehicle) (model.Vehicle, error) {
	q := `
		INSERT INTO vehicle (driver_id, nopol, vehicle_type, is_default)
		VALUES ($1::uuid, $2, $3, $4)
		ON CONFLICT (driver_id, nopol) DO UPDATE
			SET vehicle_type = EXCLUDED.vehicle_type,
			    is_default   = EXCLUDED.is_default
		RETURNING id, driver_id, nopol, vehicle_type, is_default, created_at`

	var row struct {
		ID          string    `db:"id"`
		DriverID    string    `db:"driver_id"`
		Nopol       string    `db:"nopol"`
		VehicleType string    `db:"vehicle_type"`
		IsDefault   bool      `db:"is_default"`
		CreatedAt   time.Time `db:"created_at"`
	}
	if err := r.db.QueryRowxContext(ctx, q,
		v.DriverID, v.Nopol, string(v.VehicleType), v.IsDefault,
	).StructScan(&row); err != nil {
		return model.Vehicle{}, fmt.Errorf("register vehicle: %w", err)
	}
	return model.Vehicle{
		ID:          row.ID,
		DriverID:    row.DriverID,
		Nopol:       row.Nopol,
		VehicleType: model.VehicleType(row.VehicleType),
		IsDefault:   row.IsDefault,
		CreatedAt:   row.CreatedAt,
	}, nil
}

// ListByDriverID returns all vehicles for a driver, default-first then chronological.
func (r *vehicleRepository) ListByDriverID(ctx context.Context, driverID string) ([]model.Vehicle, error) {
	q := `
		SELECT id, driver_id, nopol, vehicle_type, is_default, created_at
		FROM vehicle
		WHERE driver_id = $1::uuid
		ORDER BY is_default DESC, created_at ASC`

	rows, err := r.db.QueryxContext(ctx, q, driverID)
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []model.Vehicle
	for rows.Next() {
		var row struct {
			ID          string    `db:"id"`
			DriverID    string    `db:"driver_id"`
			Nopol       string    `db:"nopol"`
			VehicleType string    `db:"vehicle_type"`
			IsDefault   bool      `db:"is_default"`
			CreatedAt   time.Time `db:"created_at"`
		}
		if err := rows.StructScan(&row); err != nil {
			return nil, fmt.Errorf("scan vehicle row: %w", err)
		}
		vehicles = append(vehicles, model.Vehicle{
			ID:          row.ID,
			DriverID:    row.DriverID,
			Nopol:       row.Nopol,
			VehicleType: model.VehicleType(row.VehicleType),
			IsDefault:   row.IsDefault,
			CreatedAt:   row.CreatedAt,
		})
	}
	return vehicles, rows.Err()
}
