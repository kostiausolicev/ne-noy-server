package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/repository"
	"ne_noy/tests"
)

func setupServiceWithDB(t *testing.T) (*userService, *gorm.DB, func()) {
	gormDB := tests.SetupPostgres(t)
	repo := repository.NewUserRepository(gormDB)
	roleRepo := repository.NewRoleRepository(gormDB)

	svc := &userService{r: repo, rr: roleRepo}

	cleanup := func() {
		// nothing: test container cleanup handled in tests.SetupPostgres
	}
	return svc, gormDB, cleanup
}

func createTestUser(t *testing.T, svc *userService, vkId int64, first, last string) *dto.UserDto {
	ctx := context.Background()
	userDto := dto.UserDto{
		VkId:                  vkId,
		FirstName:             first,
		LastName:              last,
		PhotoURL:              "http://photo",
		GeoAvailable:          false,
		NotificationAvailable: true,
	}

	res, err := svc.CreateUser(ctx, userDto)
	require.NoError(t, err)
	require.NotNil(t, res)
	return res
}

func TestCreateAndGetUserByVkId(t *testing.T) {
	svc, _, _ := setupServiceWithDB(t)
	ctx := context.Background()

	// Создаём пользователя
	vk := int64(123456789)
	created := createTestUser(t, svc, vk, "Ivan", "Ivanov")
	require.Equal(t, vk, created.VkId)
	require.Equal(t, "Ivan", created.FirstName)
	require.Equal(t, "Ivanov", created.LastName)

	// Получаем пользователя
	got, err := svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, created.VkId, got.VkId)
}

func TestUpdatePermissionsAndRole(t *testing.T) {
	svc, _, _ := setupServiceWithDB(t)
	ctx := context.Background()

	vk := int64(222333444)
	_ = createTestUser(t, svc, vk, "Petr", "Petrov")

	// Update geo permission
	err := svc.UpdatePermissions(ctx, "geo", vk, true)
	require.NoError(t, err)

	got, err := svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.True(t, got.GeoAvailable)

	// Update notification permission
	err = svc.UpdatePermissions(ctx, "notification", vk, false)
	require.NoError(t, err)
	got, err = svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.False(t, got.NotificationAvailable)

	// Invalid permission
	err = svc.UpdatePermissions(ctx, "unknown", vk, true)
	require.Error(t, err)

	// Update role to admin (use role id from migrations)
	adminRoleId, _ := uuid.Parse("704a0251-f27d-4f54-b09b-e17f8be2d905")
	err = svc.UpdateRole(ctx, vk, adminRoleId)
	require.NoError(t, err)

	// Verify IsAdmin becomes true when role is admin
	got, err = svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.True(t, got.IsAdmin)
}

func TestGetAllUsersFiltering(t *testing.T) {
	svc, gormDB, _ := setupServiceWithDB(t)
	ctx := context.Background()

	// создаём несколько пользователей
	u1 := createTestUser(t, svc, 1001, "Anna", "Sokolova")
	createTestUser(t, svc, 1002, "Anastasia", "Ivanova")
	createTestUser(t, svc, 1003, "Ivan", "Petrov")

	// присвоим admin роль первому пользователю, чтобы поисковые запросы возвращали результаты
	adminRoleId, _ := uuid.Parse("704a0251-f27d-4f54-b09b-e17f8be2d905")
	err := svc.UpdateRole(ctx, u1.VkId, adminRoleId)
	require.NoError(t, err)

	// empty fio -> все
	all, err := svc.GetAllUsers(ctx, "")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(all), 3)

	// single name search (should include admin) — ищем Anna, т.к. ей присвоен admin
	rez, err := svc.GetAllUsers(ctx, "Anna")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rez), 1)

	// two-part name (Anna Sokolova should be found)
	rez2, err := svc.GetAllUsers(ctx, "Anna Sokolova")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rez2), 1)

	// invalid (>2 parts)
	_, err = svc.GetAllUsers(ctx, "one two three four")
	require.Error(t, err)

	_ = gormDB
}

func TestUserModelToDto_IsEduAndAdmin(t *testing.T) {
	svc, gormDB, _ := setupServiceWithDB(t)
	ctx := context.Background()

	// Создаём пользователя
	vk := int64(555666777)
	_ = createTestUser(t, svc, vk, "Edu", "Part")

	// Вставляем роль с именем config.RoleEduPart
	reRoleId := uuid.New()
	res := gormDB.Table("roles").Create(map[string]interface{}{"id": reRoleId, "name": config.RoleEduPart, "display_name": "Edu Participant"})
	require.NoError(t, res.Error)

	// Присваиваем пользователю эту роль
	err := svc.UpdateRole(ctx, vk, reRoleId)
	require.NoError(t, err)

	// Проверяем сырые модели: роль пользователя и имя роли
	userModel, err := svc.r.GetByVkId(ctx, vk)
	require.NoError(t, err)
	require.NotNil(t, userModel)
	require.NotNil(t, userModel.Role)
	require.Equal(t, config.RoleEduPart, userModel.Role.Name, "role name in DB should match config.RoleEduPart")

	got, err := svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.NotNil(t, got)
	// IsEduParticipant определяется по имени роли
	require.Equal(t, userModel.Role.Name == config.RoleEduPart, got.IsEduParticipant)

	// Ensure IsAdmin false unless admin or event_orgs exists
	require.False(t, got.IsAdmin)
}

func TestExistEventOrgAffectsIsAdmin(t *testing.T) {
	svc, gormDB, _ := setupServiceWithDB(t)
	ctx := context.Background()

	vk := int64(888999000)
	_ = createTestUser(t, svc, vk, "Org", "User")

	// By default not admin
	got, err := svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.False(t, got.IsAdmin)

	// Получаем модель пользователя
	userModel, err := svc.r.GetByVkId(ctx, vk)
	require.NoError(t, err)
	require.NotNil(t, userModel)

	// Создаём событие и вставляем запись в event_orgs
	eventID := uuid.New()
	res := gormDB.Table("events").Create(map[string]interface{}{"id": eventID, "name": "test event", "created_at": time.Now()})
	require.NoError(t, res.Error)

	res = gormDB.Table("event_orgs").Create(map[string]interface{}{"event_id": eventID, "user_id": userModel.ID})
	require.NoError(t, res.Error)

	// Now userModelToDto should detect organizer status
	got2, err := svc.GetUserByVkId(ctx, vk)
	require.NoError(t, err)
	require.True(t, got2.IsAdmin)
}
