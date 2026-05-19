package repository

import (
	"context"
	"strings"
	"sync"

	"github.com/farid/user-service/internal/user/model"
)

// MockVehicleRepository is a thread-safe in-memory fake for tests.
// Register is idempotent on (driver_id, nopol): re-registering the same plate
// updates vehicle_type / is_default and returns the current record.
type MockVehicleRepository struct {
	mu sync.Mutex
	// byDriver maps driver_id → slice of vehicles
	byDriver map[string][]model.Vehicle

	RegisterFn       func(ctx context.Context, v model.Vehicle) (model.Vehicle, error)
	ListByDriverIDFn func(ctx context.Context, driverID string) ([]model.Vehicle, error)
}

// NewMockVehicleRepository returns an empty mock with default in-memory behaviour.
func NewMockVehicleRepository() *MockVehicleRepository {
	return &MockVehicleRepository{byDriver: map[string][]model.Vehicle{}}
}

func (m *MockVehicleRepository) Register(ctx context.Context, v model.Vehicle) (model.Vehicle, error) {
	if m.RegisterFn != nil {
		return m.RegisterFn(ctx, v)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	nopol := strings.ToUpper(strings.ReplaceAll(v.Nopol, " ", ""))
	vehicles := m.byDriver[v.DriverID]

	for i, existing := range vehicles {
		if existing.Nopol == nopol {
			vehicles[i].VehicleType = v.VehicleType
			vehicles[i].IsDefault = v.IsDefault
			m.byDriver[v.DriverID] = vehicles
			return vehicles[i], nil
		}
	}

	v.Nopol = nopol
	if v.ID == "" {
		v.ID = "mock-veh-" + v.DriverID + "-" + nopol
	}
	m.byDriver[v.DriverID] = append(vehicles, v)
	return v, nil
}

func (m *MockVehicleRepository) ListByDriverID(ctx context.Context, driverID string) ([]model.Vehicle, error) {
	if m.ListByDriverIDFn != nil {
		return m.ListByDriverIDFn(ctx, driverID)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Vehicle, len(m.byDriver[driverID]))
	copy(out, m.byDriver[driverID])
	return out, nil
}
