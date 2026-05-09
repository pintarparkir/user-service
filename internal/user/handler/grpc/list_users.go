package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/model"
)

// ListUsers paginates active users with optional status filter.
func (h *userHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	out, err := h.usecase.ListUsers(ctx, model.ListUsersRequest{
		Limit:        int(req.Limit),
		Offset:       int(req.Offset),
		StatusFilter: statusFromProto(req.StatusFilter),
	})
	if err != nil {
		return nil, err
	}
	resp := &userv1.ListUsersResponse{Total: int32(out.Total)}
	for i := range out.Users {
		resp.Users = append(resp.Users, toUserProto(&out.Users[i]))
	}
	return resp, nil
}
