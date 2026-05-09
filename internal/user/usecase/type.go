package usecase

import (
	"context"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/internal/user/repository"
	"github.com/farid/user-service/pkg/redis"
)

// UserUsecase orchestrates user-domain business logic.
type UserUsecase interface {
	// CreateUser is idempotent on external_user_id.
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByExternalID(ctx context.Context, externalUserID string) (*model.User, error)
	UpdateUser(ctx context.Context, req model.UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, req model.ListUsersRequest) (*model.ListUsersResponse, error)

	// UpsertDriver is called by the gateway on every request. It creates the driver
	// profile on first contact (lazy registration) and returns the existing one on retry.
	// Keyed on ExternalUserID from JWT; MSISDN stored for SMS notification.
	UpsertDriver(ctx context.Context, req model.UpsertDriverRequest) (*model.User, error)

	// RegisterVehicle registers a license plate for a driver (idempotent on nopol).
	RegisterVehicle(ctx context.Context, req model.RegisterVehicleRequest) (*model.Vehicle, error)
	// ListVehicles returns all registered plates for a driver.
	ListVehicles(ctx context.Context, driverID string) ([]model.Vehicle, error)
}

// userUsecase wires repository + vehicleRepo + cache.
type userUsecase struct {
	repo        repository.UserRepository
	vehicleRepo repository.VehicleRepository
	cache       redis.Collections
}
