// Package usecase implements user business logic.
package usecase

import (
	"context"
	"errors"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
)

// CreateUser is idempotent on (external_user_id):
//  1. Validate input.
//  2. If a row already exists for this external id → return it (replay-safe).
//  3. Otherwise insert. If a concurrent insert wins (UNIQUE), re-fetch and return.
func (u *userUsecase) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if existing, err := u.repo.GetByExternalID(ctx, req.ExternalUserID); err == nil && existing != nil {
		return existing, nil
	} else if err != nil && !errors.Is(err, apperror.ErrNotFound) && !apperror.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	created, err := u.repo.Create(ctx, model.User{
		ExternalUserID: req.ExternalUserID,
		FullName:       req.FullName,
		PhoneE164:      req.PhoneE164,
		Email:          req.Email,
		Status:         model.UserActive,
	})
	if err != nil {
		// Race: a concurrent caller won the insert — fetch and return.
		if apperror.Is(err, apperror.ErrConflict) {
			return u.repo.GetByExternalID(ctx, req.ExternalUserID)
		}
		return nil, err
	}
	return &created, nil
}
