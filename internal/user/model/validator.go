package model

import (
	"net/mail"
	"regexp"
	"strings"

	apperror "github.com/farid/user-service/pkg/error"
)

// e164 is a permissive E.164 check (+ followed by 8–15 digits).
var e164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

// nopolRE accepts Indonesian license plates after normalisation (no spaces, uppercase).
// Format: 1-2 region letters + 1-4 digits + 0-3 suffix letters.
// Examples: B1234ABC, AA123B, D9876XY
var nopolRE = regexp.MustCompile(`^[A-Z]{1,2}[0-9]{1,4}[A-Z]{0,3}$`)

// Validate runs cheap input checks before the request reaches the DB.
// Returning *AppError so the gRPC interceptor maps it to InvalidArgument cleanly.
func (r CreateUserRequest) Validate() error {
	switch {
	case strings.TrimSpace(r.ExternalUserID) == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "external_user_id required"}
	case strings.TrimSpace(r.FullName) == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "full_name required"}
	case r.PhoneE164 != "" && !e164.MatchString(r.PhoneE164):
		return &apperror.AppError{Code: "VALIDATION", Message: "phone_e164 must be E.164 (+<digits>)"}
	case r.Email != "":
		if _, err := mail.ParseAddress(r.Email); err != nil {
			return &apperror.AppError{Code: "VALIDATION", Message: "email invalid"}
		}
	}
	return nil
}

// Validate for the update path. ID is required; other fields optional.
func (r UpdateUserRequest) Validate() error {
	switch {
	case strings.TrimSpace(r.ID) == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "id required"}
	case r.PhoneE164 != "" && !e164.MatchString(r.PhoneE164):
		return &apperror.AppError{Code: "VALIDATION", Message: "phone_e164 must be E.164"}
	case r.Email != "":
		if _, err := mail.ParseAddress(r.Email); err != nil {
			return &apperror.AppError{Code: "VALIDATION", Message: "email invalid"}
		}
	}
	return nil
}

// Validate for the upsert-driver path. MSISDN and external_user_id are required.
func (r UpsertDriverRequest) Validate() error {
	switch {
	case strings.TrimSpace(r.ExternalUserID) == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "external_user_id required"}
	case strings.TrimSpace(r.PhoneE164) == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "phone_e164 required"}
	case !e164.MatchString(r.PhoneE164):
		return &apperror.AppError{Code: "VALIDATION", Message: "phone_e164 must be E.164 (+<digits>)"}
	}
	return nil
}

// Validate for vehicle registration. Nopol is normalised (uppercased, spaces removed)
// before the regex check so "B 1234 ABC" and "b1234abc" both pass.
func (r RegisterVehicleRequest) Validate() error {
	normalised := strings.ToUpper(strings.ReplaceAll(r.Nopol, " ", ""))
	switch {
	case strings.TrimSpace(r.DriverID) == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "driver_id required"}
	case normalised == "":
		return &apperror.AppError{Code: "VALIDATION", Message: "nopol required"}
	case !nopolRE.MatchString(normalised):
		return &apperror.AppError{Code: "VALIDATION", Message: "nopol format invalid (e.g. B1234ABC)"}
	case !IsValidVehicleType(r.VehicleType):
		return &apperror.AppError{Code: "VALIDATION", Message: "vehicle_type must be CAR or MOTORCYCLE"}
	}
	return nil
}
