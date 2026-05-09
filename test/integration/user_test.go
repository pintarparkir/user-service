//go:build integration
// +build integration

// Integration test — exercises the User repository against a real Postgres
// (with pgcrypto). Run via: `make up && make test-integration`.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	userpg "github.com/farid/user-service/internal/user/repository/postgres"
	useruc "github.com/farid/user-service/internal/user/usecase"
)

func connectUserDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("postgres",
		"host=localhost port=5432 user=postgres password=postgres dbname=parkirpintar sslmode=disable")
	require.NoError(t, err)
	return db
}

func TestIntegration_User_CRUD_RoundTrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := connectUserDB(t)
	defer db.Close()

	uc := useruc.NewUserUsecase(userpg.NewUserRepository(db, "test-key"), userpg.NewVehicleRepository(db), nil)

	extID := "ext-" + uuid.NewString()[:8]

	// 1. Create.
	created, err := uc.CreateUser(ctx, model.CreateUserRequest{
		ExternalUserID: extID,
		FullName:       "Farid Test",
		PhoneE164:      "+628111000111",
		Email:          "farid+test@example.com",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	require.Equal(t, model.USER_ACTIVE, created.Status)
	t.Cleanup(func() { _ = uc.DeleteUser(ctx, created.ID) })

	// 2. PII round-trip — phone/email returned plaintext after pgcrypto decrypt.
	require.Equal(t, "+628111000111", created.PhoneE164)
	require.Equal(t, "farid+test@example.com", created.Email)

	// 3. Idempotency — second create with same external_user_id returns existing.
	again, err := uc.CreateUser(ctx, model.CreateUserRequest{
		ExternalUserID: extID, FullName: "DIFFERENT",
	})
	require.NoError(t, err)
	require.Equal(t, created.ID, again.ID, "must return existing record (idempotent)")

	// 4. Update — happy path.
	updated, err := uc.UpdateUser(ctx, model.UpdateUserRequest{
		ID: created.ID, FullName: "Farid Updated", ExpectedVersion: created.Version,
	})
	require.NoError(t, err)
	require.Equal(t, "Farid Updated", updated.FullName)
	require.Equal(t, created.Version+1, updated.Version)

	// 5. Get by id.
	fetched, err := uc.GetUserByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "Farid Updated", fetched.FullName)

	// 6. Delete (soft).
	require.NoError(t, uc.DeleteUser(ctx, created.ID))

	// 7. After delete, list should not include this user.
	listed, err := uc.ListUsers(ctx, model.ListUsersRequest{Limit: 100})
	require.NoError(t, err)
	for _, u := range listed.Users {
		require.NotEqual(t, created.ID, u.ID, "soft-deleted user must not appear in list")
	}
}
