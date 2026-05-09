package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"
)

// DeleteUser performs a soft delete.
func (h *userHandler) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	if err := h.usecase.DeleteUser(ctx, req.Id); err != nil {
		return nil, err
	}
	return &userv1.DeleteUserResponse{Ok: true}, nil
}
