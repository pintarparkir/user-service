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

// validationError is a shortcut to build a VALIDATION AppError.
func validationError(msg string) error {
	return &apperror.AppError{Code: "VALIDATION", Message: msg}
}

// validatePhone checks E.164 format when phone is provided.
func validatePhone(phone string) error {
	if phone != "" && !e164.MatchString(phone) {
		return validationError("phone_e164 must be E.164 (+<digits>)")
	}
	return nil
}

// validateEmail checks RFC 5322 format when email is provided.
func validateEmail(email string) error {
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return validationError("email invalid")
		}
	}
	return nil
}

// requireNonEmpty returns a validation error if the trimmed value is empty.
func requireNonEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return validationError(fieldName + " required")
	}
	return nil
}

// Validate runs cheap input checks before the request reaches the DB.
func (r CreateUserRequest) Validate() error {
	if err := requireNonEmpty(r.ExternalUserID, "external_user_id"); err != nil {
		return err
	}
	if err := requireNonEmpty(r.FullName, "full_name"); err != nil {
		return err
	}
	if err := validatePhone(r.PhoneE164); err != nil {
		return err
	}
	return validateEmail(r.Email)
}

// Validate for the update path. ID is required; other fields optional.
func (r UpdateUserRequest) Validate() error {
	if err := requireNonEmpty(r.ID, "id"); err != nil {
		return err
	}
	if err := validatePhone(r.PhoneE164); err != nil {
		return err
	}
	return validateEmail(r.Email)
}

// Validate for the upsert-driver path. MSISDN and external_user_id are required.
func (r UpsertDriverRequest) Validate() error {
	if err := requireNonEmpty(r.ExternalUserID, "external_user_id"); err != nil {
		return err
	}
	if err := requireNonEmpty(r.PhoneE164, "phone_e164"); err != nil {
		return err
	}
	if !e164.MatchString(r.PhoneE164) {
		return validationError("phone_e164 must be E.164 (+<digits>)")
	}
	return nil
}

// Validate for vehicle registration. Nopol is normalised (uppercased, spaces removed)
// before the regex check so "B 1234 ABC" and "b1234abc" both pass.
func (r RegisterVehicleRequest) Validate() error {
	if err := requireNonEmpty(r.DriverID, "driver_id"); err != nil {
		return err
	}
	normalised := strings.ToUpper(strings.ReplaceAll(r.Nopol, " ", ""))
	if normalised == "" {
		return validationError("nopol required")
	}
	if !nopolRE.MatchString(normalised) {
		return validationError("nopol format invalid (e.g. B1234ABC)")
	}
	if !IsValidVehicleType(r.VehicleType) {
		return validationError("vehicle_type must be CAR or MOTORCYCLE")
	}
	return nil
}
