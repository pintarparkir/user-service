package postgres

import (
	"context"
	"fmt"

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
		limit = model.DefaultListLimit
	}
	if limit > model.MaxListLimit {
		limit = model.MaxListLimit
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
	defer func() { _ = rows.Close() }()

	users := make([]model.User, 0, limit)
	for rows.Next() {
		var row userRow
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, err
		}
		users = append(users, row.toModel())
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
