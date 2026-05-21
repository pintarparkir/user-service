package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	mockcache "github.com/farid/user-service/mock/redis"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestDeleteUser_SoftDeletesUser(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-del-001", FullName: "ToDelete",
	})
	require.NoError(t, err)

	require.NoError(t, uc.DeleteUser(context.Background(), created.ID))

	user, err := repo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, model.UserDeleted, user.Status, "status should be DELETED after soft delete")
}

func TestDeleteUser_Idempotent(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-del-002", FullName: "ToDelete",
	})
	require.NoError(t, err)

	require.NoError(t, uc.DeleteUser(context.Background(), created.ID))
	// Second call must also succeed — callers can safely retry.
	require.NoError(t, uc.DeleteUser(context.Background(), created.ID))
}

func TestDeleteUser_NonExistentID_IsIdempotent(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	// Deleting an ID that never existed should not return an error.
	require.NoError(t, uc.DeleteUser(context.Background(), "non-existent-id"))
}

func TestDeleteUser_InvalidatesCache(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	cache := mockcache.NewMockCache()
	uc := NewUserUsecase(repo, nil, cache)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-del-003", FullName: "Cached",
	})
	require.NoError(t, err)

	// Pre-seed cache as if a previous GetUserByID had cached the user.
	cacheKey := "user:" + created.ID
	cache.Store[cacheKey] = `{"id":"` + created.ID + `","full_name":"Cached"}`

	require.NoError(t, uc.DeleteUser(context.Background(), created.ID))

	_, exists := cache.Store[cacheKey]
	require.False(t, exists, "cache entry should be removed after delete")
}
