package postgres

import (
	"github.com/jmoiron/sqlx"

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
