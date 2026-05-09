package grpc

import (
	"context"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/model"
)

// RegisterVehicle registers a license plate for a driver.
// Idempotent on (driver_id, nopol): re-registering returns the current record.
func (h *userHandler) RegisterVehicle(ctx context.Context, req *userv1.RegisterVehicleRequest) (*userv1.Vehicle, error) {
	out, err := h.usecase.RegisterVehicle(ctx, model.RegisterVehicleRequest{
		DriverID:    req.DriverId,
		Nopol:       req.Nopol,
		VehicleType: vehicleTypeFromProto(req.VehicleType),
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	return toVehicleProto(out), nil
}
