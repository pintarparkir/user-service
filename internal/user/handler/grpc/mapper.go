package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/model"
)

// toUserProto converts the domain aggregate into protobuf wire format.
func toUserProto(u *model.User) *userv1.User {
	if u == nil {
		return nil
	}
	return &userv1.User{
		Id:             u.ID,
		ExternalUserId: u.ExternalUserID,
		FullName:       u.FullName,
		PhoneE164:      u.PhoneE164,
		Email:          u.Email,
		Status:         statusToProto(u.Status),
		Version:        int32(u.Version),
		CreatedAt:      timestamppb.New(u.CreatedAt),
		UpdatedAt:      timestamppb.New(u.UpdatedAt),
	}
}

// toVehicleProto converts a domain Vehicle to its proto representation.
func toVehicleProto(v *model.Vehicle) *userv1.Vehicle {
	if v == nil {
		return nil
	}
	return &userv1.Vehicle{
		Id:          v.ID,
		DriverId:    v.DriverID,
		Nopol:       v.Nopol,
		VehicleType: vehicleTypeToProto(v.VehicleType),
		IsDefault:   v.IsDefault,
		CreatedAt:   timestamppb.New(v.CreatedAt),
	}
}

func statusToProto(s model.UserStatus) userv1.UserStatus {
	switch s {
	case model.USER_ACTIVE:
		return userv1.UserStatus_USER_STATUS_ACTIVE
	case model.USER_SUSPENDED:
		return userv1.UserStatus_USER_STATUS_SUSPENDED
	case model.USER_DELETED:
		return userv1.UserStatus_USER_STATUS_DELETED
	}
	return userv1.UserStatus_USER_STATUS_UNSPECIFIED
}

func statusFromProto(s userv1.UserStatus) model.UserStatus {
	switch s {
	case userv1.UserStatus_USER_STATUS_ACTIVE:
		return model.USER_ACTIVE
	case userv1.UserStatus_USER_STATUS_SUSPENDED:
		return model.USER_SUSPENDED
	case userv1.UserStatus_USER_STATUS_DELETED:
		return model.USER_DELETED
	}
	return ""
}

func vehicleTypeToProto(vt model.VehicleType) userv1.VehicleType {
	switch vt {
	case model.VehicleTypeCar:
		return userv1.VehicleType_VEHICLE_TYPE_CAR
	case model.VehicleTypeMotorcycle:
		return userv1.VehicleType_VEHICLE_TYPE_MOTORCYCLE
	}
	return userv1.VehicleType_VEHICLE_TYPE_UNSPECIFIED
}

func vehicleTypeFromProto(vt userv1.VehicleType) model.VehicleType {
	switch vt {
	case userv1.VehicleType_VEHICLE_TYPE_CAR:
		return model.VehicleTypeCar
	case userv1.VehicleType_VEHICLE_TYPE_MOTORCYCLE:
		return model.VehicleTypeMotorcycle
	}
	return ""
}
