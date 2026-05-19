package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestListVehicles_EmptyDriverID_ValidationError(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	_, err := uc.ListVehicles(context.Background(), "")
	require.Error(t, err)
}

func TestListVehicles_NoRegisteredVehicles_ReturnsEmptySlice(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	driver, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-listveh-001", FullName: "Driver",
	})
	require.NoError(t, err)

	vehicles, err := uc.ListVehicles(context.Background(), driver.ID)
	require.NoError(t, err)
	require.NotNil(t, vehicles, "should return an empty slice, not nil")
	require.Empty(t, vehicles)
}

func TestListVehicles_ReturnsAllVehiclesForDriver(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	driver, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-listveh-002", FullName: "Driver",
	})
	require.NoError(t, err)

	plates := []string{"B1234ABC", "D5678XY"}
	for _, nopol := range plates {
		_, err = uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
			DriverID:    driver.ID,
			Nopol:       nopol,
			VehicleType: model.VehicleTypeCar,
		})
		require.NoError(t, err)
	}

	vehicles, err := uc.ListVehicles(context.Background(), driver.ID)
	require.NoError(t, err)
	require.Len(t, vehicles, 2)
}

func TestListVehicles_OnlyReturnsVehiclesForRequestedDriver(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	vehicleRepo := mockrepo.NewMockVehicleRepository()
	uc := NewUserUsecase(repo, vehicleRepo, nil)

	driverA, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-listveh-003a", FullName: "DriverA",
	})
	require.NoError(t, err)
	driverB, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-listveh-003b", FullName: "DriverB",
	})
	require.NoError(t, err)

	_, err = uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID: driverA.ID, Nopol: "B1111AA", VehicleType: model.VehicleTypeCar,
	})
	require.NoError(t, err)
	_, err = uc.RegisterVehicle(context.Background(), model.RegisterVehicleRequest{
		DriverID: driverB.ID, Nopol: "D2222BB", VehicleType: model.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)

	vehiclesA, err := uc.ListVehicles(context.Background(), driverA.ID)
	require.NoError(t, err)
	require.Len(t, vehiclesA, 1)
	require.Equal(t, driverA.ID, vehiclesA[0].DriverID)
}
