// Package model defines user domain models and constants.
package model

// UserStatus represents the account state.
type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserSuspended UserStatus = "SUSPENDED"
	UserDeleted   UserStatus = "DELETED"
)

// IsValidStatus reports whether s is a recognised value.
func IsValidStatus(s UserStatus) bool {
	switch s {
	case UserActive, UserSuspended, UserDeleted:
		return true
	}
	return false
}

// Idempotency / RPC scopes — matches FullMethod produced by gRPC code-gen.
const (
	ScopeCreateUser      = "/parkirpintar.user.v1.UserService/CreateUser"
	ScopeUpdateUser      = "/parkirpintar.user.v1.UserService/UpdateUser"
	ScopeUpsertDriver    = "/parkirpintar.user.v1.UserService/UpsertDriver"
	ScopeRegisterVehicle = "/parkirpintar.user.v1.UserService/RegisterVehicle"
)

// Pagination defaults.
const (
	DefaultListLimit = 50
	MaxListLimit     = 200
)
