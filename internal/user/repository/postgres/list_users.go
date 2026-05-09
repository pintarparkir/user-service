package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/farid/user-service/internal/user/model"
)

// List returns paginated users with optional status filter, plus the unfiltered
// total (without limit/offset) for client pagination UI.
//
// Tradeoff note: the COUNT(*) is a second query; for very large tables we'd
// switch to estimated counts via pg_class.reltuples.
func (r *userRepository) List(ctx context.Context, req model.ListUsersRequest) ([]model.User, int, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = model.DEFAULT_LIST_LIMIT
	}
	if limit > model.MAX_LIST_LIMIT {
		limit = model.MAX_LIST_LIMIT
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	where := `WHERE status != 'DELETED'`
	args := []interface{}{r.pgKey, r.pgKey, limit, offset}
	if req.StatusFilter != "" {
		where += ` AND status = $5`
		args = append(args, req.StatusFilter)
	}

	listQ := fmt.Sprintf(`
		SELECT %s FROM user_profile %s
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		fmt.Sprintf(columns, 1, 2), where,
	)

	rows, err := r.db.QueryxContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]model.User, 0, limit)
	for rows.Next() {
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
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, err
		}
		users = append(users, model.User{
			ID: row.ID, ExternalUserID: row.ExternalUserID, FullName: row.FullName,
			PhoneE164: row.PhoneE164, Email: row.Email,
			Status: model.UserStatus(row.Status), Version: row.Version,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
	}

	// Total (unfiltered by limit/offset, but respects status filter).
	countWhere := `WHERE status != 'DELETED'`
	countArgs := []interface{}{}
	if req.StatusFilter != "" {
		countWhere += ` AND status = $1`
		countArgs = append(countArgs, req.StatusFilter)
	}
	var total int
	if err := r.db.QueryRowxContext(ctx, `SELECT COUNT(*) FROM user_profile `+countWhere, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	return users, total, nil
}
