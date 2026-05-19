package http

import (
	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/pkg/utils"
)

func (h *userHandler) listVehicles(c *gin.Context) {
	vehicles, err := h.usecase.ListVehicles(c.Request.Context(), c.GetString(ctxDriverID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	dtos := make([]*vehicleDTO, 0, len(vehicles))
	for i := range vehicles {
		dtos = append(dtos, toVehicleDTO(&vehicles[i]))
	}

	utils.OK(c, gin.H{"vehicles": dtos}, "vehicles retrieved successfully")
}
