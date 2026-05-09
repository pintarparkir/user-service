package http

import (
	"time"

	"github.com/farid/user-service/internal/user/model"
)

type profileDTO struct {
	ID             string    `json:"id"`
	ExternalUserID string    `json:"external_user_id"`
	FullName       string    `json:"full_name"`
	PhoneE164      string    `json:"phone_e164,omitempty"`
	Email          string    `json:"email,omitempty"`
	Status         string    `json:"status"`
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func toProfileDTO(u *model.User) *profileDTO {
	if u == nil {
		return nil
	}
	return &profileDTO{
		ID: u.ID, ExternalUserID: u.ExternalUserID, FullName: u.FullName,
		PhoneE164: u.PhoneE164, Email: u.Email,
		Status: string(u.Status), Version: u.Version,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

type updateMeReq struct {
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	ExpectedVersion int    `json:"expected_version"`
}

type vehicleDTO struct {
	ID          string    `json:"id"`
	Nopol       string    `json:"nopol"`
	VehicleType string    `json:"vehicle_type"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

func toVehicleDTO(v *model.Vehicle) *vehicleDTO {
	return &vehicleDTO{
		ID:          v.ID,
		Nopol:       v.Nopol,
		VehicleType: string(v.VehicleType),
		IsDefault:   v.IsDefault,
		CreatedAt:   v.CreatedAt,
	}
}

type registerVehicleReq struct {
	Nopol       string `json:"nopol"        binding:"required"`
	VehicleType string `json:"vehicle_type" binding:"required"`
	IsDefault   bool   `json:"is_default"`
}
