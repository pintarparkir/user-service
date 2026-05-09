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

// GetByID returns the user or ErrNotFound. PII is decrypted inline.
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	q := fmt.Sprintf(`SELECT %s FROM user_profile WHERE id = $3`, fmt.Sprintf(columns, 1, 2))
	return r.scanOne(ctx, q, r.pgKey, r.pgKey, id)
}

// GetByExternalID returns the user keyed by super-app identity.
func (r *userRepository) GetByExternalID(ctx context.Context, externalUserID string) (*model.User, error) {
	q := fmt.Sprintf(`SELECT %s FROM user_profile WHERE external_user_id = $3`, fmt.Sprintf(columns, 1, 2))
	return r.scanOne(ctx, q, r.pgKey, r.pgKey, externalUserID)
}

// scanOne factors the SELECT-and-decrypt boilerplate.
func (r *userRepository) scanOne(ctx context.Context, q string, args ...interface{}) (*model.User, error) {
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
	if err := r.db.QueryRowxContext(ctx, q, args...).StructScan(&row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	return &model.User{
		ID: row.ID, ExternalUserID: row.ExternalUserID, FullName: row.FullName,
		PhoneE164: row.PhoneE164, Email: row.Email,
		Status: model.UserStatus(row.Status), Version: row.Version,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}
