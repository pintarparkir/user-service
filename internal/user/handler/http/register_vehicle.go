package http

import (
	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/pkg/utils"
)

func (h *userHandler) registerVehicle(c *gin.Context) {
	var body registerVehicleReq
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, err)
		return
	}

	vehicle, err := h.usecase.RegisterVehicle(c.Request.Context(), model.RegisterVehicleRequest{
		DriverID:    c.GetString(ctxDriverID),
		Nopol:       body.Nopol,
		VehicleType: model.VehicleType(body.VehicleType),
		IsDefault:   body.IsDefault,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Created(c, toVehicleDTO(vehicle), "vehicle registered successfully")
}
