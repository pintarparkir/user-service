// Package postgres implements user repository using PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
)

// Create inserts a new user. PII columns are encrypted via pgp_sym_encrypt.
// Returns ErrConflict on duplicate external_user_id (idempotency replay should
// be handled in the usecase by calling GetByExternalID first).
func (r *userRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	q := `
		INSERT INTO user_profile (
			id, external_user_id, full_name,
			phone_e164_enc, email_enc, status, version
		) VALUES (
			COALESCE(NULLIF($1,''), gen_random_uuid()::text)::uuid,
			$2, $3,
			NULLIF(pgp_sym_encrypt($4, $7), pgp_sym_encrypt('', $7)),
			NULLIF(pgp_sym_encrypt($5, $7), pgp_sym_encrypt('', $7)),
			$6, 1
		)
		RETURNING id, external_user_id, full_name,
		          COALESCE(pgp_sym_decrypt(phone_e164_enc, $7), '') AS phone_e164,
		          COALESCE(pgp_sym_decrypt(email_enc,      $7), '') AS email,
		          status, version, created_at, updated_at`

	var row userRow
	err := r.db.QueryRowxContext(ctx, q,
		u.ID, u.ExternalUserID, u.FullName, u.PhoneE164, u.Email, u.Status, r.pgKey,
	).StructScan(&row)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && string(pgErr.Code) == "23505" {
			return model.User{}, apperror.ErrConflict
		}
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, apperror.ErrNotFound
		}
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}
	return row.toModel(), nil
}
