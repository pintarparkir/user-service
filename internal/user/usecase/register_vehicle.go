package usecase

import (
	"context"
	"strings"

	"github.com/farid/user-service/internal/user/model"
)

// RegisterVehicle registers a license plate for a driver.
// Nopol is normalised (upper-cased, spaces stripped) before storage.
// Idempotent on (driver_id, nopol): re-registering the same plate updates
// vehicle_type / is_default and returns the current record.
func (u *userUsecase) RegisterVehicle(ctx context.Context, req model.RegisterVehicleRequest) (*model.Vehicle, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify driver exists before writing vehicle.
	if _, err := u.repo.GetByID(ctx, req.DriverID); err != nil {
		return nil, err
	}

	v, err := u.vehicleRepo.Register(ctx, model.Vehicle{
		DriverID:    req.DriverID,
		Nopol:       strings.ToUpper(strings.ReplaceAll(req.Nopol, " ", "")),
		VehicleType: req.VehicleType,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	return &v, nil
}
