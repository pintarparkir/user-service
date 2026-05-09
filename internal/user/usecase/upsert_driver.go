package usecase

import (
	"context"

	"github.com/farid/user-service/internal/user/model"
)

// UpsertDriver is the lazy-registration entry point called by the gateway on
// every inbound request. It is safe to call repeatedly — if a profile already
// exists for the given external_user_id it is returned unchanged.
func (u *userUsecase) UpsertDriver(ctx context.Context, req model.UpsertDriverRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return u.repo.GetOrCreateByMSISDN(ctx, req.PhoneE164, req.ExternalUserID, req.FullName)
}
