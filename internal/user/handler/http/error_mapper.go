package http

import (
	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/pkg/utils"
)

// renderError is a convenience wrapper around utils.Error for backward compatibility.
// Handlers should prefer using utils.Error directly.
func renderError(c *gin.Context, err error) {
	utils.Error(c, err)
}
