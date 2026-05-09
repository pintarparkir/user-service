package usecase

import (
	"github.com/farid/user-service/internal/user/repository"
	"github.com/farid/user-service/pkg/redis"
)

// NewUserUsecase wires the usecase. Cache is optional (pass nil to disable
// the per-id cache); usecase will skip cache ops gracefully.
func NewUserUsecase(repo repository.UserRepository, vehicleRepo repository.VehicleRepository, cache redis.Collections) UserUsecase {
	return &userUsecase{repo: repo, vehicleRepo: vehicleRepo, cache: cache}
}
