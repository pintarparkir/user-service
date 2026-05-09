package model

// ListUsersResponse pairs a page of results with the unfiltered total
// so the client can render pagination controls without an extra round-trip.
type ListUsersResponse struct {
	Users []User
	Total int
}

// ListVehiclesResponse returns all vehicles registered for a driver.
type ListVehiclesResponse struct {
	Vehicles []Vehicle
}
