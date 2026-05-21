package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"
)

// GetUserByID hits the cache-first read path.
func (h *userHandler) GetUserByID(ctx context.Context, req *userv1.GetUserByIdRequest) (*userv1.User, error) {
	out, err := h.usecase.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toUserProto(out), nil
}

// GetUserByExternalID is used at SSO sign-in.
func (h *userHandler) GetUserByExternalID(ctx context.Context, req *userv1.GetUserByExternalIdRequest) (*userv1.User, error) {
	out, err := h.usecase.GetUserByExternalID(ctx, req.ExternalUserId)
	if err != nil {
		return nil, err
	}
	return toUserProto(out), nil
}
