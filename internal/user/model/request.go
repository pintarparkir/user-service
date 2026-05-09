package model

// CreateUserRequest is the domain-level shape for new user creation.
// Idempotency is enforced at the repository layer via UNIQUE on external_user_id.
type CreateUserRequest struct {
	ExternalUserID string
	FullName       string
	PhoneE164      string
	Email          string
}

// UpsertDriverRequest creates or returns the driver profile keyed on MSISDN.
// Called by the gateway on every request when the JWT carries a new user.
type UpsertDriverRequest struct {
	PhoneE164      string // from JWT claim (required)
	ExternalUserID string // super-app identity (required)
	FullName       string // optional — updated on subsequent calls
}

// RegisterVehicleRequest registers a license plate under a driver.
// Idempotent on (driver_id, nopol): re-registering the same plate is a no-op.
type RegisterVehicleRequest struct {
	DriverID    string
	Nopol       string // e.g. "B1234ABC"
	VehicleType VehicleType
	IsDefault   bool
}

// ListVehiclesRequest lists all vehicles registered for a driver.
type ListVehiclesRequest struct {
	DriverID string
}

// UpdateUserRequest carries optional patch fields. Empty string = "do not change"
// (we do not need to support clearing strings in this MVP).
type UpdateUserRequest struct {
	ID              string
	FullName        string
	PhoneE164       string
	Email           string
	ExpectedVersion int
}

// ListUsersRequest carries pagination + status filter.
type ListUsersRequest struct {
	Limit        int
	Offset       int
	StatusFilter UserStatus // empty = no filter
}
