package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
)

// Update applies a partial patch with optimistic locking on `version`.
// Empty string in u.FullName/PhoneE164/Email = "do not change that column".
// Returns ErrConflict if version check fails (concurrent edit).
func (r *userRepository) Update(ctx context.Context, u model.User, expectedVersion int) (model.User, error) {
	q := `
		UPDATE user_profile SET
			full_name      = COALESCE(NULLIF($1, ''), full_name),
			phone_e164_enc = CASE WHEN $2 = '' THEN phone_e164_enc
			                      ELSE pgp_sym_encrypt($2, $5) END,
			email_enc      = CASE WHEN $3 = '' THEN email_enc
			                      ELSE pgp_sym_encrypt($3, $5) END,
			version        = version + 1,
			updated_at     = now()
		WHERE id = $4 AND version = $6 AND status != 'DELETED'
		RETURNING id, external_user_id, full_name,
		          COALESCE(pgp_sym_decrypt(phone_e164_enc, $5), '') AS phone_e164,
		          COALESCE(pgp_sym_decrypt(email_enc,      $5), '') AS email,
		          status, version, created_at, updated_at`

	var row struct {
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
	err := r.db.QueryRowxContext(ctx, q,
		u.FullName, u.PhoneE164, u.Email, u.ID, r.pgKey, expectedVersion,
	).StructScan(&row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, apperror.ErrConflict
	}
	if err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}
	return model.User{
		ID: row.ID, ExternalUserID: row.ExternalUserID, FullName: row.FullName,
		PhoneE164: row.PhoneE164, Email: row.Email,
		Status: model.UserStatus(row.Status), Version: row.Version,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}
