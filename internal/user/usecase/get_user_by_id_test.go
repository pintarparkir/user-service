package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
	mockcache "github.com/farid/user-service/mock/redis"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestGetUserByID_Found(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-get-001", FullName: "Farid",
	})
	require.NoError(t, err)

	got, err := uc.GetUserByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "Farid", got.FullName)
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.GetUserByID(context.Background(), "does-not-exist")
	require.Error(t, err)
	require.True(t, apperror.Is(err, apperror.ErrNotFound))
}

func TestGetUserByID_CacheMiss_PopulatesCache(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	cache := mockcache.NewMockCache()
	uc := NewUserUsecase(repo, nil, cache)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-get-002", FullName: "Alice",
	})
	require.NoError(t, err)

	got, err := uc.GetUserByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	// After the first call, the cache should be populated.
	_, exists := cache.Store["user:"+created.ID]
	require.True(t, exists, "cache should be populated after a cache-miss read")
}

func TestGetUserByID_CacheHit_SkipsRepo(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	cache := mockcache.NewMockCache()
	uc := NewUserUsecase(repo, nil, cache)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-get-003", FullName: "Bob",
	})
	require.NoError(t, err)

	// First call: populates cache.
	_, err = uc.GetUserByID(context.Background(), created.ID)
	require.NoError(t, err)

	// Break the repo so any subsequent repo call would panic.
	repo.GetByIDFn = func(_ context.Context, _ string) (*model.User, error) {
		t.Fatal("repo.GetByID should not be called on a cache hit")
		return nil, nil
	}

	// Second call: must be served from cache without hitting the repo.
	got, err := uc.GetUserByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
}

func TestGetUserByExternalID_Found(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-get-004", FullName: "Charlie",
	})
	require.NoError(t, err)

	got, err := uc.GetUserByExternalID(context.Background(), "ext-get-004")
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)
}

func TestGetUserByExternalID_NotFound(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.GetUserByExternalID(context.Background(), "unknown-ext-id")
	require.Error(t, err)
	require.True(t, apperror.Is(err, apperror.ErrNotFound))
}
