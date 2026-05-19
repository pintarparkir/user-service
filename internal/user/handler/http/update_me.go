package http

import (
	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/pkg/utils"
)

// PUT /v1/me — update own profile (name, email). Phone is managed by super-app.
func (h *userHandler) updateMe(c *gin.Context) {
	var body updateMeReq
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, err)
		return
	}

	out, err := h.usecase.UpdateUser(c.Request.Context(), model.UpdateUserRequest{
		ID:              c.GetString(ctxDriverID),
		FullName:        body.FullName,
		Email:           body.Email,
		ExpectedVersion: body.ExpectedVersion,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.OK(c, toProfileDTO(out), "profile updated successfully")
}
