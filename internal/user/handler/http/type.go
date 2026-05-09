// Package http exposes the User domain over REST/JSON for the Tencent Mini Program.
//
// Routes are mounted at /v1. JWT verification + lazy driver registration are handled
// by middleware before each handler runs.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/farid/user-service/internal/user/usecase"
)

const (
	ctxDriverID = "driver_id"
)

type userHandler struct {
	usecase usecase.UserUsecase
	jwtKey  string // super-app RS256 public key PEM; empty = skip verify (dev)
}

// RegisterUserHandler mounts mini-app REST routes under rg.
// jwtPubKeyPEM is the super-app RS256 public key; pass "" to skip sig verification in dev.
func RegisterUserHandler(rg *gin.RouterGroup, uc usecase.UserUsecase, jwtPubKeyPEM string) {
	h := &userHandler{usecase: uc, jwtKey: jwtPubKeyPEM}

	authed := rg.Group("")
	authed.Use(h.jwtMiddleware(), h.upsertDriverMiddleware())

	authed.GET("/me", h.getMe)
	authed.PUT("/me", h.updateMe)
	authed.GET("/me/vehicles", h.listVehicles)
	authed.POST("/me/vehicles", h.registerVehicle)
}
