package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestListUsers_Empty(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	resp, err := uc.ListUsers(context.Background(), model.ListUsersRequest{})
	require.NoError(t, err)
	require.Empty(t, resp.Users)
	require.Equal(t, 0, resp.Total)
}

func TestListUsers_ReturnsAllActiveUsers(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	for _, name := range []string{"Alice", "Bob", "Charlie"} {
		_, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
			ExternalUserID: "ext-list-" + name, FullName: name,
		})
		require.NoError(t, err)
	}

	resp, err := uc.ListUsers(context.Background(), model.ListUsersRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, resp.Total)
	require.Len(t, resp.Users, 3)
}

func TestListUsers_ExcludesDeletedUsers(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	created, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-list-del", FullName: "DeleteMe",
	})
	require.NoError(t, err)
	_, err = uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-list-keep", FullName: "KeepMe",
	})
	require.NoError(t, err)

	require.NoError(t, uc.DeleteUser(context.Background(), created.ID))

	resp, err := uc.ListUsers(context.Background(), model.ListUsersRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, resp.Total, "deleted users must be excluded from the default list")
}

func TestListUsers_StatusFilter(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	active, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-list-active", FullName: "ActiveUser",
	})
	require.NoError(t, err)
	require.NoError(t, uc.DeleteUser(context.Background(), active.ID))

	// Filtering by DELETED should return the deleted user only.
	resp, err := uc.ListUsers(context.Background(), model.ListUsersRequest{
		StatusFilter: model.USER_DELETED,
	})
	require.NoError(t, err)
	require.Equal(t, 1, resp.Total)
	require.Equal(t, model.USER_DELETED, resp.Users[0].Status)
}
