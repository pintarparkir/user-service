package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	mockrepo "github.com/farid/user-service/mock/repository"
)

func TestUpsertDriver_CreatesNewDriverOnFirstContact(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	got, err := uc.UpsertDriver(context.Background(), model.UpsertDriverRequest{
		ExternalUserID: "ext-upsert-001",
		PhoneE164:      "+628111111111",
		FullName:       "Budi",
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotEmpty(t, got.ID)
	require.Equal(t, model.UserActive, got.Status)
}

func TestUpsertDriver_Idempotent_ReturnsSameUser(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	req := model.UpsertDriverRequest{
		ExternalUserID: "ext-upsert-002",
		PhoneE164:      "+628222222222",
		FullName:       "Dewi",
	}

	first, err := uc.UpsertDriver(context.Background(), req)
	require.NoError(t, err)

	// Calling again with the same external_user_id must return the existing profile.
	second, err := uc.UpsertDriver(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID, "repeated upsert must return the same user ID")
}

func TestUpsertDriver_Validation_MissingExternalID(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.UpsertDriver(context.Background(), model.UpsertDriverRequest{
		PhoneE164: "+628111111111",
	})
	require.Error(t, err)
}

func TestUpsertDriver_Validation_MissingPhone(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.UpsertDriver(context.Background(), model.UpsertDriverRequest{
		ExternalUserID: "ext-upsert-003",
	})
	require.Error(t, err)
}

func TestUpsertDriver_Validation_InvalidPhoneFormat(t *testing.T) {
	repo := mockrepo.NewMockUserRepository()
	uc := NewUserUsecase(repo, nil, nil)

	_, err := uc.UpsertDriver(context.Background(), model.UpsertDriverRequest{
		ExternalUserID: "ext-upsert-004",
		PhoneE164:      "08111111111", // missing leading +
	})
	require.Error(t, err)
}
