package usecase

import (
	"context"

	"github.com/farid/user-service/internal/user/model"
)

// UpdateUser applies a partial patch. Empty string in any field = unchanged.
// Optimistic-lock check happens at the repository (returns ErrConflict on stale).
func (u *userUsecase) UpdateUser(ctx context.Context, req model.UpdateUserRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	updated, err := u.repo.Update(ctx, model.User{
		ID:        req.ID,
		FullName:  req.FullName,
		PhoneE164: req.PhoneE164,
		Email:     req.Email,
	}, req.ExpectedVersion)
	if err != nil {
		return nil, err
	}
	if u.cache != nil {
		_ = u.cache.Del(ctx, "user:"+updated.ID)
	}
	return &updated, nil
}
