package repository

import (
	"context"

	"github.com/farid/user-service/internal/user/model"
)

// UserRepository is the storage port. Adapters live under repository/postgres.
//
// Note: pgcrypto encrypt/decrypt for PII columns happens inside the adapter
// to keep the usecase free of infrastructure concerns. The encryption key is
// passed at adapter construction time.
type UserRepository interface {
	Create(ctx context.Context, u model.User) (model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByExternalID(ctx context.Context, externalUserID string) (*model.User, error)
	// GetOrCreateByMSISDN is the lazy-registration entry point for mini-app users.
	// It upserts a driver profile keyed on external_user_id and stores the MSISDN
	// for SMS notification. Safe to call on every request from the gateway.
	GetOrCreateByMSISDN(ctx context.Context, msisdn, externalUserID, fullName string) (*model.User, error)
	Update(ctx context.Context, u model.User, expectedVersion int) (model.User, error)
	SoftDelete(ctx context.Context, id string) error
	List(ctx context.Context, req model.ListUsersRequest) ([]model.User, int, error)
}

// VehicleRepository manages the vehicle plate registry for each driver.
type VehicleRepository interface {
	// Register inserts or updates a vehicle plate. Idempotent on (driver_id, nopol).
	Register(ctx context.Context, v model.Vehicle) (model.Vehicle, error)
	ListByDriverID(ctx context.Context, driverID string) ([]model.Vehicle, error)
}
