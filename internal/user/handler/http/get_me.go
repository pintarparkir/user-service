package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *userHandler) getMe(c *gin.Context) {
	user, err := h.usecase.GetUserByID(c.Request.Context(), c.GetString(ctxDriverID))
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, toProfileDTO(user))
}
