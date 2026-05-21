// Package grpc implements gRPC handlers for user service.
package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/model"
)

// CreateUser is idempotent on `external_user_id` — see usecase.CreateUser.
func (h *userHandler) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.User, error) {
	out, err := h.usecase.CreateUser(ctx, model.CreateUserRequest{
		ExternalUserID: req.ExternalUserId,
		FullName:       req.FullName,
		PhoneE164:      req.PhoneE164,
		Email:          req.Email,
	})
	if err != nil {
		return nil, err
	}
	return toUserProto(out), nil
}
