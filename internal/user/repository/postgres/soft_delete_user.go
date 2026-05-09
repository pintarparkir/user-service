package postgres

import (
	"context"
	"database/sql"
	"errors"

	apperror "github.com/farid/user-service/pkg/error"
)

// SoftDelete flips status to 'DELETED' so the row stays for audit.
// Idempotent — calling on an already-deleted user is a no-op (returns nil).
func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE user_profile SET status='DELETED', version=version+1, updated_at=now()
		 WHERE id = $1 AND status != 'DELETED'`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Distinguish "doesn't exist" from "already deleted".
		var exists bool
		qErr := r.db.QueryRowxContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_profile WHERE id=$1)`, id).Scan(&exists)
		if errors.Is(qErr, sql.ErrNoRows) || !exists {
			return apperror.ErrNotFound
		}
		// Already deleted → idempotent success.
	}
	return nil
}
