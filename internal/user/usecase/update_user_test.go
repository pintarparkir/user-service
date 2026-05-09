package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apperror "github.com/farid/user-service/pkg/error"

	"github.com/farid/user-service/internal/user/model"
	mockrepo "github.com/farid/user-service/mock/repository"
)

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
