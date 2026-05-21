package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/farid/user-service/internal/user/model"
	"github.com/farid/user-service/internal/user/repository/postgres"
	apperror "github.com/farid/user-service/pkg/error"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return sqlx.NewDb(db, "postgres"), mock
}

func userRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "external_user_id", "full_name", "phone_e164", "email", "status", "version", "created_at", "updated_at",
	}).AddRow(
		"user-1", "ext-1", "John Doe", "628123456789", "john@example.com", "ACTIVE", 1, time.Now().UTC(), time.Now().UTC(),
	)
}

func TestUserRepo_Create_HappyPath(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`INSERT INTO user_profile`).
		WithArgs("", "ext-1", "John Doe", "628123456789", "john@example.com", model.UserActive, "test-key").
		WillReturnRows(userRows())

	got, err := repo.Create(ctx, model.User{
		ExternalUserID: "ext-1",
		FullName:       "John Doe",
		PhoneE164:      "628123456789",
		Email:          "john@example.com",
		Status:         model.UserActive,
	})

	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.Equal(t, "John Doe", got.FullName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Create_UniqueViolation(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`INSERT INTO user_profile`).
		WillReturnError(&pq.Error{Code: "23505"})

	_, err := repo.Create(ctx, model.User{ExternalUserID: "ext-1", FullName: "John", Status: model.UserActive})

	require.Error(t, err)
	assert.True(t, apperror.Is(err, apperror.ErrConflict))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE id`).
		WithArgs("test-key", "test-key", "user-1").
		WillReturnRows(userRows())

	got, err := repo.GetByID(ctx, "user-1")

	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.Equal(t, "John Doe", got.FullName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE id`).
		WithArgs("test-key", "test-key", "missing").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(ctx, "missing")

	require.Error(t, err)
	assert.True(t, apperror.Is(err, apperror.ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByExternalID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE external_user_id`).
		WithArgs("test-key", "test-key", "ext-1").
		WillReturnRows(userRows())

	got, err := repo.GetByExternalID(ctx, "ext-1")

	require.NoError(t, err)
	assert.Equal(t, "ext-1", got.ExternalUserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetOrCreateByMSISDN_ExistingUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE external_user_id`).
		WithArgs("test-key", "test-key", "ext-1").
		WillReturnRows(userRows())

	got, err := repo.GetOrCreateByMSISDN(ctx, "628123456789", "ext-1", "John Doe")

	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetOrCreateByMSISDN_CreateNew(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE external_user_id`).
		WithArgs("test-key", "test-key", "ext-new").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO user_profile`).
		WithArgs("", "ext-new", "Driver", "628999", "", model.UserActive, "test-key").
		WillReturnRows(userRows())

	got, err := repo.GetOrCreateByMSISDN(ctx, "628999", "ext-new", "")

	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetOrCreateByMSISDN_RaceRecovery(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE external_user_id`).
		WithArgs("test-key", "test-key", "ext-race").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO user_profile`).
		WillReturnError(&pq.Error{Code: "23505"})
	mock.ExpectQuery(`SELECT .+ FROM user_profile WHERE external_user_id`).
		WithArgs("test-key", "test-key", "ext-race").
		WillReturnRows(userRows())

	got, err := repo.GetOrCreateByMSISDN(ctx, "628999", "ext-race", "Driver")

	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Update_HappyPath(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`UPDATE user_profile SET`).
		WithArgs("Jane Doe", "628777", "jane@example.com", "user-1", "test-key", 1).
		WillReturnRows(userRows())

	got, err := repo.Update(ctx, model.User{
		ID: "user-1", FullName: "Jane Doe", PhoneE164: "628777", Email: "jane@example.com",
	}, 1)

	require.NoError(t, err)
	assert.Equal(t, "user-1", got.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Update_Conflict(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`UPDATE user_profile SET`).
		WithArgs("Jane Doe", "", "", "user-1", "test-key", 99).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.Update(ctx, model.User{ID: "user-1", FullName: "Jane Doe"}, 99)

	require.Error(t, err)
	assert.True(t, apperror.Is(err, apperror.ErrConflict))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_List_HappyPath(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectQuery(`SELECT .* FROM user_profile`).
		WithArgs("test-key", "test-key", 10, 0).
		WillReturnRows(userRows())
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_profile`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	users, total, err := repo.List(ctx, model.ListUsersRequest{Limit: 10})

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, users, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_SoftDelete_HappyPath(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectExec(`UPDATE user_profile SET status='DELETED'`).
		WithArgs("user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SoftDelete(ctx, "user-1")

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_SoftDelete_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewUserRepository(db, "test-key")

	mock.ExpectExec(`UPDATE user_profile SET status='DELETED'`).
		WithArgs("missing").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := repo.SoftDelete(ctx, "missing")

	require.Error(t, err)
	assert.True(t, apperror.Is(err, apperror.ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVehicleRepo_Register_HappyPath(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewVehicleRepository(db)

	mock.ExpectQuery(`INSERT INTO vehicle`).
		WithArgs("user-1", "B1234XYZ", string(model.VehicleTypeCar), true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "driver_id", "nopol", "vehicle_type", "is_default", "created_at"}).
			AddRow("veh-1", "user-1", "B1234XYZ", "CAR", true, time.Now().UTC()))

	got, err := repo.Register(ctx, model.Vehicle{DriverID: "user-1", Nopol: "B1234XYZ", VehicleType: model.VehicleTypeCar, IsDefault: true})

	require.NoError(t, err)
	assert.Equal(t, "veh-1", got.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVehicleRepo_ListByDriverID_HappyPath(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := postgres.NewVehicleRepository(db)

	mock.ExpectQuery(`SELECT id, driver_id, nopol`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "driver_id", "nopol", "vehicle_type", "is_default", "created_at"}).
			AddRow("veh-1", "user-1", "B1234XYZ", "CAR", true, time.Now().UTC()))

	vehicles, err := repo.ListByDriverID(ctx, "user-1")

	require.NoError(t, err)
	assert.Len(t, vehicles, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
