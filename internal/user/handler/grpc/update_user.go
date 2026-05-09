package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/model"
)

// UpdateUser applies a partial patch with optimistic-lock check.
func (h *userHandler) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.User, error) {
	out, err := h.usecase.UpdateUser(ctx, model.UpdateUserRequest{
		ID:              req.Id,
		FullName:        req.FullName,
		PhoneE164:       req.PhoneE164,
		Email:           req.Email,
		ExpectedVersion: int(req.ExpectedVersion),
	})
	if err != nil {
		return nil, err
	}
	return toUserProto(out), nil
}
