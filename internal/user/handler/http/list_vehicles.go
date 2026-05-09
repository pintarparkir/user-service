package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *userHandler) listVehicles(c *gin.Context) {
	vehicles, err := h.usecase.ListVehicles(c.Request.Context(), c.GetString(ctxDriverID))
	if err != nil {
		renderError(c, err)
		return
	}

	dtos := make([]*vehicleDTO, 0, len(vehicles))
	for i := range vehicles {
		dtos = append(dtos, toVehicleDTO(&vehicles[i]))
	}

	c.JSON(http.StatusOK, gin.H{"vehicles": dtos})
}
