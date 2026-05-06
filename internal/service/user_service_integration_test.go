package service

import (
	"context"
	"database/sql"
	"testing"

	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	repopgx "ne_noy/internal/repository/impl"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubVkClient struct{}

func (stubVkClient) GetVkUsers(userIds []string) ([]dto.CreateUserDto, error) {
	return nil, nil
}

func (stubVkClient) SendNotification(userIds []string, messageText, fragment string) (dto.SendMessageResponse, error) {
	return dto.SendMessageResponse{}, nil
}

func TestUserServiceCreateUser(t *testing.T) {
	// Test case: создание пользователя с ролью по умолчанию и сохранением записи в БД
	// Input: пустая тестовая БД после миграций и CreateUserDto{VkId:101, FirstName:"Ivan", LastName:"Petrov", PhotoURL:"https://example.com/avatar.jpg"}
	// Expected: сервис возвращает созданного пользователя с ролью default, а в таблице users появляется запись с тем же vk_id и полями профиля
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	roleRepo := repopgx.NewRoleRepositoryPgx(pool)
	userRepo := repopgx.NewUserRepository(pool)
	svc := NewUserService(userRepo, roleRepo, stubVkClient{})

	result, err := svc.CreateUser(context.Background(), dto.CreateUserDto{
		VkId:      101,
		FirstName: "Ivan",
		LastName:  "Petrov",
		PhotoURL:  "https://example.com/avatar.jpg",
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, int64(101), result.VkId)
	assert.Equal(t, "Ivan", result.FirstName)
	assert.Equal(t, "Petrov", result.LastName)
	assert.Equal(t, config.RoleDefault, result.Role.Name)
	assert.False(t, result.GeoAvailable)
	assert.False(t, result.NotificationAvailable)

	var (
		dbID                  uuid.UUID
		firstName             string
		lastName              string
		photoURL              string
		roleName              string
		geoAvailable          bool
		notificationAvailable bool
	)
	err = sqlDB.QueryRow(`
		SELECT u.id, u.first_name, u.last_name, u.photo_url, r.name, u.geo_available, u.notification_available
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.vk_id = $1
	`, 101).Scan(&dbID, &firstName, &lastName, &photoURL, &roleName, &geoAvailable, &notificationAvailable)
	require.NoError(t, err)

	assert.Equal(t, dbID, *result.ID)
	assert.Equal(t, "Ivan", firstName)
	assert.Equal(t, "Petrov", lastName)
	assert.Equal(t, "https://example.com/avatar.jpg", photoURL)
	assert.Equal(t, config.RoleDefault, roleName)
	assert.False(t, geoAvailable)
	assert.False(t, notificationAvailable)
}

func TestUserServiceUpdatePermissions(t *testing.T) {
	// Test case: обновление флага geo_available у существующего пользователя
	// Input: в БД есть пользователь с vk_id=202 и geo_available=false, вызов UpdatePermissions("geo", 202, true)
	// Expected: сервис возвращает nil, а в таблице users поле geo_available становится true
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	userID := insertUser(t, sqlDB, 202, "Anna", "Smirnova", config.RoleDefault)
	require.NotEqual(t, uuid.Nil, userID)

	svc := NewUserService(repopgx.NewUserRepository(pool), repopgx.NewRoleRepositoryPgx(pool), stubVkClient{})

	err := svc.UpdatePermissions(context.Background(), "geo", 202, true)
	require.NoError(t, err)

	var geoAvailable bool
	err = sqlDB.QueryRow(`SELECT geo_available FROM users WHERE vk_id = $1`, 202).Scan(&geoAvailable)
	require.NoError(t, err)
	assert.True(t, geoAvailable)
}

func TestUserServiceGetUserByVkId(t *testing.T) {
	// Test case: получение существующего пользователя по vk_id
	// Input: в БД есть пользователь с vk_id=303, ролью default и без участия в event_orgs
	// Expected: сервис возвращает UserDto с данными пользователя, ролью default и признаком IsAdmin=false
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	userID := insertUser(t, sqlDB, 303, "Petr", "Sidorov", config.RoleDefault)
	require.NotEqual(t, uuid.Nil, userID)

	svc := NewUserService(repopgx.NewUserRepository(pool), repopgx.NewRoleRepositoryPgx(pool), stubVkClient{})

	result, err := svc.GetUserByVkId(context.Background(), 303)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, userID, *result.ID)
	assert.Equal(t, int64(303), result.VkId)
	assert.Equal(t, "Petr", result.FirstName)
	assert.Equal(t, "Sidorov", result.LastName)
	assert.Equal(t, config.RoleDefault, result.Role.Name)
	assert.False(t, result.IsAdmin)
}

func insertUser(t *testing.T, db *sql.DB, vkID int64, firstName, lastName, roleName string) uuid.UUID {
	t.Helper()

	roleID := getRoleID(t, db, roleName)
	userID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO users (id, vk_id, first_name, last_name, role_id, photo_url, geo_available, notification_available)
		VALUES ($1, $2, $3, $4, $5, $6, false, false)
	`, userID, vkID, firstName, lastName, roleID, "https://example.com/test.jpg")
	require.NoError(t, err)

	return userID
}

func getRoleID(t *testing.T, db *sql.DB, roleName string) uuid.UUID {
	t.Helper()

	var roleID uuid.UUID
	err := db.QueryRow(`SELECT id FROM roles WHERE name = $1`, roleName).Scan(&roleID)
	require.NoError(t, err)
	return roleID
}

func insertUserWithPhoto(t *testing.T, db *sql.DB, vkID int64, firstName, lastName, roleName, photoURL string) uuid.UUID {
	t.Helper()

	roleID := getRoleID(t, db, roleName)
	userID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO users (id, vk_id, first_name, last_name, role_id, photo_url, geo_available, notification_available)
		VALUES ($1, $2, $3, $4, $5, $6, false, false)
	`, userID, vkID, firstName, lastName, roleID, photoURL)
	require.NoError(t, err)

	return userID
}
