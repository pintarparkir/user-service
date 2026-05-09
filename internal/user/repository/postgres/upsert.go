package postgres

import (
	"context"
	"errors"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
)

// GetOrCreateByMSISDN performs a read-or-create keyed on external_user_id.
// If the driver already exists (same external_user_id) it is returned as-is.
// If not, a new profile is created with the supplied MSISDN as phone_e164.
// On a concurrent insert race (UNIQUE violation), it re-fetches and returns.
func (r *userRepository) GetOrCreateByMSISDN(ctx context.Context, msisdn, externalUserID, fullName string) (*model.User, error) {
	if existing, err := r.GetByExternalID(ctx, externalUserID); err == nil {
		return existing, nil
	} else if !errors.Is(err, apperror.ErrNotFound) && !apperror.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	name := fullName
	if name == "" {
		name = "Driver"
	}
	created, err := r.Create(ctx, model.User{
		ExternalUserID: externalUserID,
		FullName:       name,
		PhoneE164:      msisdn,
		Status:         model.USER_ACTIVE,
	})
	if err != nil {
		if apperror.Is(err, apperror.ErrConflict) {
			return r.GetByExternalID(ctx, externalUserID)
		}
		return nil, err
	}
	return &created, nil
}
