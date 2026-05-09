package usecase

import (
	"context"

	"github.com/farid/user-service/internal/user/model"
)

// ListUsers returns paginated users with optional status filter.
// Defaults & caps applied at the repository.
func (u *userUsecase) ListUsers(ctx context.Context, req model.ListUsersRequest) (*model.ListUsersResponse, error) {
	users, total, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return &model.ListUsersResponse{Users: users, Total: total}, nil
}
