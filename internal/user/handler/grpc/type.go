package grpc

import (
	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/usecase"
)

// userHandler adapts gRPC requests to the usecase layer.
type userHandler struct {
	userv1.UnimplementedUserServiceServer
	usecase usecase.UserUsecase
}
