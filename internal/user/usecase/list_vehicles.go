package usecase

import (
	"context"
	"strings"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
)

// ListVehicles returns all license plates registered for a driver.
// Returns an empty slice (not an error) when the driver has no registered vehicles.
func (u *userUsecase) ListVehicles(ctx context.Context, driverID string) ([]model.Vehicle, error) {
	if strings.TrimSpace(driverID) == "" {
		return nil, &apperror.AppError{Code: "VALIDATION", Message: "driver_id required"}
	}
	vehicles, err := u.vehicleRepo.ListByDriverID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if vehicles == nil {
		return []model.Vehicle{}, nil
	}
	return vehicles, nil
}
