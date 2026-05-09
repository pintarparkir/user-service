package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/model"
	pkgjwt "github.com/farid/user-service/pkg/jwt"
)

const (
	ctxExternalUserID = "external_user_id"
	ctxPhoneE164      = "phone_e164"
)

// jwtMiddleware parses and verifies the Bearer JWT from Authorization header.
// Sets external_user_id and phone_e164 in the Gin context.
func (h *userHandler) jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHENTICATED", "message": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(raw, "Bearer ")

		claims, err := pkgjwt.Parse(token, h.jwtKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHENTICATED", "message": err.Error()})
			return
		}

		c.Set(ctxExternalUserID, claims.Sub)
		c.Set(ctxPhoneE164, claims.Phone)
		c.Next()
	}
}

// upsertDriverMiddleware lazily registers the driver and stores driver_id in context.
// Must run after jwtMiddleware.
func (h *userHandler) upsertDriverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		extID, _ := c.Get(ctxExternalUserID)
		phone, _ := c.Get(ctxPhoneE164)

		user, err := h.usecase.UpsertDriver(c.Request.Context(), model.UpsertDriverRequest{
			ExternalUserID: extID.(string),
			PhoneE164:      phone.(string),
		})
		if err != nil {
			renderError(c, err)
			c.Abort()
			return
		}

		c.Set(ctxDriverID, user.ID)
		c.Next()
	}
}
