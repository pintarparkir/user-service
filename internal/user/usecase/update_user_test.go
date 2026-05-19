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

func TestUpdateUser_Validation_MissingID(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.UpdateUser(context.Background(), model.UpdateUserRequest{
		FullName: "SomeName",
	})
	require.Error(t, err, "update without ID must fail validation")
}

func TestUpdateUser_Validation_InvalidPhone(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.UpdateUser(context.Background(), model.UpdateUserRequest{
		ID: "any-id", PhoneE164: "0811bad",
	})
	require.Error(t, err)
}

func TestUpdateUser_PartialPatch_OnlyUpdatesProvidedFields(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-upd-partial", FullName: "Original", PhoneE164: "+628100000000",
	})
	require.NoError(t, err)

	updated, err := uc.UpdateUser(context.Background(), model.UpdateUserRequest{
		ID: created.ID, FullName: "Changed", ExpectedVersion: created.Version,
	})
	require.NoError(t, err)
	require.Equal(t, "Changed", updated.FullName)
	// Phone should be unchanged because it was not provided in the patch.
	require.Equal(t, "+628100000000", updated.PhoneE164)
}

func TestUpdateUser_InvalidatesCache(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	cache := mockcache.NewMockCache()
	uc := NewUserUsecase(repo, nil, cache)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-upd-cache", FullName: "Cached",
	})
	require.NoError(t, err)

	cache.Store["user:"+created.ID] = `{"id":"` + created.ID + `","full_name":"Cached"}`

	_, err = uc.UpdateUser(context.Background(), model.UpdateUserRequest{
		ID: created.ID, FullName: "Updated", ExpectedVersion: created.Version,
	})
	require.NoError(t, err)

	_, exists := cache.Store["user:"+created.ID]
	require.False(t, exists, "cache entry should be invalidated after update")
}

func TestUpdateUser_OptimisticLock(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)
	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-100", FullName: "v1",
	})
	require.NoError(t, err)

	// Stale version → ErrConflict.
	_, err = uc.UpdateUser(context.Background(), model.UpdateUserRequest{
		ID: created.ID, FullName: "v2", ExpectedVersion: 99,
	})
	require.Error(t, err)
	require.True(t, apperror.Is(err, apperror.ErrConflict), "expected ErrConflict on stale version")

	// Correct version → succeeds.
	updated, err := uc.UpdateUser(context.Background(), model.UpdateUserRequest{
		ID: created.ID, FullName: "v2", ExpectedVersion: created.Version,
	})
	require.NoError(t, err)
	require.Equal(t, "v2", updated.FullName)
	require.Equal(t, created.Version+1, updated.Version)
}
