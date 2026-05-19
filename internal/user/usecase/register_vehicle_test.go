package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestRegisterVehicle_HappyPath(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	driver, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-veh-001", FullName: "Driver",
	})
	require.NoError(t, err)

	v, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID:    driver.ID,
		Nopol:       "B1234ABC",
		VehicleType: model.VehicleTypeCar,
		IsDefault:   true,
	})
	require.NoError(t, err)
	require.NotNil(t, v)
	require.Equal(t, "B1234ABC", v.Nopol)
	require.Equal(t, model.VehicleTypeCar, v.VehicleType)
	require.True(t, v.IsDefault)
}

func TestRegisterVehicle_NopolNormalized(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	driver, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-veh-002", FullName: "Driver",
	})
	require.NoError(t, err)

	// Input with spaces and lowercase — should be normalised to "B1234ABC".
	v, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID:    driver.ID,
		Nopol:       "b 1234 abc",
		VehicleType: model.VehicleTypeCar,
	})
	require.NoError(t, err)
	require.Equal(t, "B1234ABC", v.Nopol, "nopol must be upper-cased with spaces stripped")
}

func TestRegisterVehicle_Idempotent_OnSameNopol(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	driver, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-veh-003", FullName: "Driver",
	})
	require.NoError(t, err)

	req := model.RegisterVehicleRequest{
		DriverID: driver.ID, Nopol: "D5678XY", VehicleType: model.VehicleTypeMotorcycle,
	}
	first, err := uc.RegisterVehicle(context.Background(), req)
	require.NoError(t, err)

	second, err := uc.RegisterVehicle(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID, "re-registering the same plate must return the existing record")
}

func TestRegisterVehicle_Validation_MissingDriverID(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	_, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		Nopol: "B1234ABC", VehicleType: model.VehicleTypeCar,
	})
	require.Error(t, err)
}

func TestRegisterVehicle_Validation_MissingNopol(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	_, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID: "some-id", VehicleType: model.VehicleTypeCar,
	})
	require.Error(t, err)
}

func TestRegisterVehicle_Validation_InvalidNopolFormat(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	_, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID: "some-id", Nopol: "12345INVALID", VehicleType: model.VehicleTypeCar,
	})
	require.Error(t, err)
}

func TestRegisterVehicle_Validation_InvalidVehicleType(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	_, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID: "some-id", Nopol: "B1234ABC", VehicleType: "TRUCK",
	})
	require.Error(t, err)
}

func TestRegisterVehicle_DriverNotFound_ReturnsError(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	_, err := uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID:    "non-existent-driver",
		Nopol:       "B1234ABC",
		VehicleType: model.VehicleTypeCar,
	})
	require.Error(t, err)
	require.True(t, apperror.Is(err, apperror.ErrNotFound), "should return ErrNotFound when driver does not exist")
}
