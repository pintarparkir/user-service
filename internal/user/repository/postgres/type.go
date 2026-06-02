package postgres

import (
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/internal/user/repository"
)

// userRepository is the Postgres adapter for the UserRepository port.
//
// pgKey is the symmetric key used by pgcrypto to encrypt PII columns.
// In production it comes from Cloud Secret Manager (see pkg/configs).
type userRepository struct {
	db    *sqlx.DB
	pgKey string
}

// NewUserRepository wires the Postgres adapter.
func NewUserRepository(db *sqlx.DB, pgcryptoKey string) repository.UserRepository {
	return &userRepository{db: db, pgKey: pgcryptoKey}
}

// columns is the canonical SELECT projection. PII columns are decrypted inline
// using pgcrypto's pgp_sym_decrypt, returning plaintext to the application.
const columns = `
	id,
	external_user_id,
	full_name,
	COALESCE(pgp_sym_decrypt(phone_e164_enc, $%d), '') AS phone_e164,
	COALESCE(pgp_sym_decrypt(email_enc,      $%d), '') AS email,
	status,
	version,
	created_at,
	updated_at`

// userRow is the shared scan target for all user SELECT queries.
// Centralised here to eliminate row-struct duplication across repository files.
type userRow struct {
	ID             string    `db:"id"`
	ExternalUserID string    `db:"external_user_id"`
	FullName       string    `db:"full_name"`
	PhoneE164      string    `db:"phone_e164"`
	Email          string    `db:"email"`
	Status         string    `db:"status"`
	Version        int       `db:"version"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// toModel converts the DB row into a domain model.
func (r userRow) toModel() model.User {
	return model.User{
		ID:             r.ID,
		ExternalUserID: r.ExternalUserID,
		FullName:       r.FullName,
		PhoneE164:      r.PhoneE164,
		Email:          r.Email,
		Status:         model.UserStatus(r.Status),
		Version:        r.Version,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// vehicleRow is the shared scan target for vehicle SELECT queries.
type vehicleRow struct {
	ID          string    `db:"id"`
	DriverID    string    `db:"driver_id"`
	Nopol       string    `db:"nopol"`
	VehicleType string    `db:"vehicle_type"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
}

// toModel converts the vehicle DB row into a domain model.
func (r vehicleRow) toModel() model.Vehicle {
	return model.Vehicle{
		ID:          r.ID,
		DriverID:    r.DriverID,
		Nopol:       r.Nopol,
		VehicleType: model.VehicleType(r.VehicleType),
		IsDefault:   r.IsDefault,
		CreatedAt:   r.CreatedAt,
	}
}
