package grpc

import (
	"google.golang.org/grpc"

	userv1 "github.com/farid/user-service/api/proto/user/v1"

	"github.com/farid/user-service/internal/user/usecase"
)

// RegisterUserHandler attaches the user gRPC handler to the server.
func RegisterUserHandler(server *grpc.Server, uc usecase.UserUsecase) {
	userv1.RegisterUserServiceServer(server, &userHandler{usecase: uc})
}
