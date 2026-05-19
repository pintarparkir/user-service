package http

import (
	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/pkg/utils"
)

func (h *userHandler) getMe(c *gin.Context) {
	user, err := h.usecase.GetUserByID(c.Request.Context(), c.GetString(ctxDriverID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.OK(c, toProfileDTO(user), "profile retrieved successfully")
}
