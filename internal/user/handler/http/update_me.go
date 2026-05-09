package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/model"
)

// PUT /v1/me — update own profile (name, email). Phone is managed by super-app.
func (h *userHandler) updateMe(c *gin.Context) {
	var body updateMeReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "BAD_REQUEST", "message": err.Error()})
		return
	}

	out, err := h.usecase.UpdateUser(c.Request.Context(), model.UpdateUserRequest{
		ID:              c.GetString(ctxDriverID),
		FullName:        body.FullName,
		Email:           body.Email,
		ExpectedVersion: body.ExpectedVersion,
	})
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, toProfileDTO(out))
}
