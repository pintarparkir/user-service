package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/model"
)

func (h *userHandler) registerVehicle(c *gin.Context) {
	var body registerVehicleReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "BAD_REQUEST", "message": err.Error()})
		return
	}

	vehicle, err := h.usecase.RegisterVehicle(c.Request.Context(), model.RegisterVehicleRequest{
		DriverID:    c.GetString(ctxDriverID),
		Nopol:       body.Nopol,
		VehicleType: model.VehicleType(body.VehicleType),
		IsDefault:   body.IsDefault,
	})
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toVehicleDTO(vehicle))
}
