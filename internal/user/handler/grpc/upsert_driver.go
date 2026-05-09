package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/model"
)

// UpsertDriver is the gateway entry point for lazy driver registration.
// Called on every request once the JWT has been verified; idempotent on MSISDN.
func (h *userHandler) UpsertDriver(ctx context.Context, req *userv1.UpsertDriverRequest) (*userv1.User, error) {
	out, err := h.usecase.UpsertDriver(ctx, model.UpsertDriverRequest{
		PhoneE164:      req.PhoneE164,
		ExternalUserID: req.ExternalUserId,
		FullName:       req.FullName,
	})
	if err != nil {
		return nil, err
	}
	return toUserProto(out), nil
}
