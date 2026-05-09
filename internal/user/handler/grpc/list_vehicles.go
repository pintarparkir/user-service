package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"
)

// ListVehicles returns all license plates registered for a driver.
func (h *userHandler) ListVehicles(ctx context.Context, req *userv1.ListVehiclesRequest) (*userv1.ListVehiclesResponse, error) {
	vehicles, err := h.usecase.ListVehicles(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}
	protos := make([]*userv1.Vehicle, 0, len(vehicles))
	for i := range vehicles {
		protos = append(protos, toVehicleProto(&vehicles[i]))
	}
	return &userv1.ListVehiclesResponse{Vehicles: protos}, nil
}
