package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"
	"ne_noy/tests"
)

// Интеграционный тест: создаёт Event в реальной БД и затем получает его через сервис.
// Входные данные: модель Event с заполненным ID, Name, StartsAt, Status.
// Ожидаем: Create через репозиторий проходит без ошибки, EventService.GetEvent возвращает DTO с тем же ID.
func TestEventService_CreateAndGetById_Integration(t *testing.T) {
	// Setup real postgres test container with migrations
	db := tests.SetupPostgres(t)
	ctx := context.Background()

	eRepo := repository.NewEventRepository(db)
	// user service not needed for Create
	svc := NewEventService(eRepo, nil)

	startsAt := time.Now().Add(24 * time.Hour)
	inDto := model.Event{
		ID:   uuid.New(),
		Name: "Integration Event",
		// cover and other fields optional
		StartsAt: &startsAt,
		Status:   ptrStatus("PUBLISH"),
	}

	created, err := eRepo.Create(ctx, &inDto)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Now use service GetEvent (requires repo GetById)
	got, err := svc.GetEvent(ctx, created.ID, 0)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, created.ID, got.ID)
}

// Интеграционный тест: проверяет обновление события через EventService.UpdateEvent.
// Сценарий: создаём пользователя (организатора) и событие, затем вызываем UpdateEvent с DTO,
// в котором указаны новый Title, список организаторов (ID пользователя) и доступные роли (UUID роли из миграций).
// Ожидаем: UpdateEvent возвращает DTO с новым Title; в БД появится запись в event_orgs и event_roles.
func TestEventService_UpdateEvent_Integration(t *testing.T) {
	// Setup DB
	db := tests.SetupPostgres(t)
	ctx := context.Background()

	userRepo := repository.NewUserRepository(db)
	eRepo := repository.NewEventRepository(db)
	svc := NewEventService(eRepo, nil)

	// create a user to be organizer
	user := model.User{
		ID:        uuid.New(),
		FirstName: "OrgFirst",
		LastName:  "OrgLast",
		VkID:      9999,
	}
	require.NoError(t, userRepo.Create(ctx, &user))

	// create initial event
	startsAt := time.Now().Add(48 * time.Hour)
	status := "DRAFT"
	evt := model.Event{
		ID:       uuid.New(),
		Name:     "Old name",
		StartsAt: &startsAt,
		Status:   &status,
	}

	created, err := eRepo.Create(ctx, &evt)
	require.NoError(t, err)

	// prepare update DTO: change title, add organizer and available role (use 'default' role id from migrations)
	roleID, _ := uuid.Parse("9ef01b95-6d87-4115-80df-7085a647bf36")
	updateDto := dto.CreateUpdateEventDto{
		Title:          ptrString("New title"),
		Orgs:           []uuid.UUID{user.ID},
		AvailableRoles: []uuid.UUID{roleID},
	}

	res, err := svc.UpdateEvent(ctx, created.ID, updateDto)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "New title", res.Title)

	// verify organizer present via repository helper
	orgs, err := eRepo.GetEventOrgs(ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, orgs, 1)
	require.Equal(t, user.ID, orgs[0].ID)

	// verify role linked via raw table
	var count int64
	db.WithContext(ctx).Table("event_roles").Where("event_id = ? AND role_id = ?", created.ID, roleID).Count(&count)
	require.Equal(t, int64(1), count)
}

func ptrStatus(s string) *string { return &s }

// ptrString defined in unit tests
