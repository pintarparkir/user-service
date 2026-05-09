package model

import "time"

// VehicleType is the category of vehicle registered to a driver.
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "CAR"
	VehicleTypeMotorcycle VehicleType = "MOTORCYCLE"
)

// IsValidVehicleType reports whether vt is a known vehicle category.
func IsValidVehicleType(vt VehicleType) bool {
	return vt == VehicleTypeCar || vt == VehicleTypeMotorcycle
}

// Vehicle is a registered license plate associated with a driver.
// One driver may register multiple vehicles; at most one has is_default = true.
type Vehicle struct {
	ID          string      `db:"id"`
	DriverID    string      `db:"driver_id"`
	Nopol       string      `db:"nopol"`
	VehicleType VehicleType `db:"vehicle_type"`
	IsDefault   bool        `db:"is_default"`
	CreatedAt   time.Time   `db:"created_at"`
}
