package model

// UserStatus represents the account state.
type UserStatus string

const (
	USER_ACTIVE    UserStatus = "ACTIVE"
	USER_SUSPENDED UserStatus = "SUSPENDED"
	USER_DELETED   UserStatus = "DELETED"
)

// IsValidStatus reports whether s is a recognised value.
func IsValidStatus(s UserStatus) bool {
	switch s {
	case USER_ACTIVE, USER_SUSPENDED, USER_DELETED:
		return true
	}
	return false
}

// Idempotency / RPC scopes — matches FullMethod produced by gRPC code-gen.
const (
	SCOPE_CREATE_USER      = "/parkirpintar.user.v1.UserService/CreateUser"
	SCOPE_UPDATE_USER      = "/parkirpintar.user.v1.UserService/UpdateUser"
	SCOPE_UPSERT_DRIVER    = "/parkirpintar.user.v1.UserService/UpsertDriver"
	SCOPE_REGISTER_VEHICLE = "/parkirpintar.user.v1.UserService/RegisterVehicle"
)

// Pagination defaults.
const (
	DEFAULT_LIST_LIMIT = 50
	MAX_LIST_LIMIT     = 200
)
