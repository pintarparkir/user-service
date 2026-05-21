package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestCreateUser_HappyPath(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	got, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-001",
		FullName:       "Farid",
		PhoneE164:      "+628111111111",
		Email:          "farid@example.com",
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "Farid", got.FullName)
	require.Equal(t, model.UserActive, got.Status)
	require.Equal(t, 1, got.Version)
}

func TestCreateUser_Idempotent_OnExternalID(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	first, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-002", FullName: "Alice",
	})
	require.NoError(t, err)

	// Second call with the same external_user_id should return the same record.
	second, err := uc.CreateUser(context.Background(), model.CreateUserRequest{
		ExternalUserID: "ext-002", FullName: "DIFFERENT NAME",
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID, "should return existing user, not insert a new one")
	require.Equal(t, "Alice", second.FullName, "should return original name (idempotent replay)")
}

func TestCreateUser_Validation(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	cases := []struct {
		name string
		req  model.CreateUserRequest
	}{
		{"missing external id", model.CreateUserRequest{FullName: "x"}},
		{"missing full name", model.CreateUserRequest{ExternalUserID: "x"}},
		{"bad phone", model.CreateUserRequest{ExternalUserID: "x", FullName: "y", PhoneE164: "0811"}},
		{"bad email", model.CreateUserRequest{ExternalUserID: "x", FullName: "y", Email: "not-an-email"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.CreateUser(context.Background(), tc.req)
			require.Error(t, err)
		})
	}
}
