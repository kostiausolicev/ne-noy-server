package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/tests/mocks"
)

// Тест проверяет успешное создание события через EventService.CreateEvent.
// Входные данные: валидный CreateUpdateEventDto с Title, Description и StartsAt.
// Ожидаем: отсутствие ошибки и DTO с Title равным входному.
func TestCreateEvent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)

	ctx := context.Background()
	startsAt := time.Now()
	status := "DRAFT"

	inDto := dto.CreateUpdateEventDto{
		Title:          ptrString("Test event"),
		Description:    ptrString("desc"),
		StartsAt:       &startsAt,
		Status:         &status,
		Orgs:           []uuid.UUID{},
		AvailableRoles: []uuid.UUID{},
	}

	// Мокаем репозиторий: Create возвращает тот же объект без ошибки
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, e *model.Event) (*model.Event, error) {
		return e, nil
	})

	res, err := svc.CreateEvent(ctx, inDto)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "Test event", res.Title)
}

// Тест проверяет поведение CreateEvent при ошибке репозитория.
// Входные данные: любой CreateUpdateEventDto, поведение репозитория: возвращает ошибку.
// Ожидаем: сервис возвращает ошибку и nil результат.
func TestCreateEvent_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	inDto := dto.CreateUpdateEventDto{Title: ptrString("X")}
	res, err := svc.CreateEvent(ctx, inDto)
	require.Error(t, err)
	require.Nil(t, res)
}

// Тест проверяет успешное обновление события через EventService.UpdateEvent.
// Входные данные: eventId и CreateUpdateEventDto с изменённым Title.
// Мок: репозиторий возвращает обновлённый model.Event.
// Ожидаем: сервис возвращает DTO с обновлённым Title без ошибок.
func TestUpdateEvent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	id := uuid.New()

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*model.Event, error) {
			// формируем корректный ответ репозитория: используем id и поля из fields
			ev := &model.Event{ID: id}
			if nameVal, ok := fields["name"]; ok {
				if nameStr, ok2 := nameVal.(string); ok2 {
					ev.Name = nameStr
				} else if p, ok3 := nameVal.(*string); ok3 && p != nil {
					ev.Name = *p
				}
			}
			if ev.StartsAt == nil {
				now := time.Now()
				ev.StartsAt = &now
			}
			if ev.Status == nil {
				s := "ACTIVE"
				ev.Status = &s
			}
			return ev, nil
		})

	inDto := dto.CreateUpdateEventDto{Title: ptrString("Updated")}
	res, err := svc.UpdateEvent(ctx, id, inDto)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "Updated", res.Title)
}

// Тест UpdateEvent при ошибке репозитория (например, event не найден).
// Ожидаем: ошибка от сервиса и nil результат.
func TestUpdateEvent_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))
	res, err := svc.UpdateEvent(ctx, uuid.New(), dto.CreateUpdateEventDto{Title: ptrString("x")})
	require.Error(t, err)
	require.Nil(t, res)
}

// Тест успешного получения подробного Event через EventService.GetEvent.
// Входные данные: существующий id события и vkId пользователя.
// Моки: repo.GetById возвращает модель события с организаторами/участниками/вложениями,
// repo.GetUserParticipationInEvent возвращает флаг участия текущего пользователя.
// Ожидаем: корректный dto.EventDto с заполненными полями и CurrentUserIsParticipant=true.
func TestGetEvent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	id := uuid.New()
	starts := time.Now()
	status := "PUBLISH"
	org := model.User{ID: uuid.New(), FirstName: "OrgF", LastName: "OrgL", VkID: 111, PhotoURL: "p"}
	participantUser := model.User{ID: uuid.New(), FirstName: "PU", LastName: "PL", VkID: 222, PhotoURL: "pp"}
	e := &model.Event{
		ID:                id,
		Name:              "EventName",
		StartsAt:          &starts,
		Status:            &status,
		Orgs:              []model.User{org},
		EventParticipants: []model.EventParticipant{{User: participantUser}},
		Attachments:       []model.EventAttachment{{AttachmentLink: "http://a"}},
	}

	mockRepo.EXPECT().GetById(gomock.Any(), id).Return(e, nil)
	mockRepo.EXPECT().GetUserParticipationInEvent(gomock.Any(), id, int64(999)).Return(true, nil)

	res, err := svc.GetEvent(ctx, id, 999)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "EventName", res.Title)
	require.Equal(t, 1, len(res.Orgs))
	require.Equal(t, 1, len(res.Attachments))
	if res.CurrentUserIsParticipant == nil {
		t.Fatalf("expected CurrentUserIsParticipant not nil")
	}
	require.True(t, *res.CurrentUserIsParticipant)
}

// Тест GetEvent, когда репозиторий возвращает ошибку.
// Ожидаем: ошибка от сервиса и nil результат.
func TestGetEvent_NotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	mockRepo.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(nil, errors.New("db err"))
	res, err := svc.GetEvent(ctx, uuid.New(), 0)
	require.Error(t, err)
	require.Nil(t, res)
}

// Тест GetEventParticipants: мокаем репозиторий, возвращаем список участников.
// Вход: id события. Ожидаем: корректный список dto.EventParticipantDto с флагом IsChecked=true.
func TestGetEventParticipants_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	id := uuid.New()
	participants := []model.EventParticipant{{User: model.User{ID: uuid.New(), FirstName: "A", LastName: "B", VkID: 1, PhotoURL: "x"}, IsChecked: true}}
	mockRepo.EXPECT().GetParticipants(gomock.Any(), id).Return(participants, nil)

	res, err := svc.GetEventParticipants(ctx, id)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.True(t, res[0].IsChecked)
}

// Тест GetEventsByRole/GetArchive: проверяем, что сервис корректно парсит события из репозитория.
// Вход: имя роли (string), мок репозитория возвращает список model.Event.
// Ожидаем: список dto.EventMiniDto той же длины.
func TestGetEventsByRoleAndArchive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockEventRepository(ctrl)
	mockUserSvc := mocks.NewMockUserService(ctrl)

	svc := NewEventService(mockRepo, mockUserSvc)
	ctx := context.Background()

	roleName := "default"
	starts := time.Now().Add(48 * time.Hour)
	status := "PUBLISH"
	mockRepo.EXPECT().GetAllByRole(gomock.Any(), roleName).Return([]*model.Event{{ID: uuid.New(), Name: "R1", StartsAt: &starts, Status: &status}}, nil)
	mockRepo.EXPECT().GetAllArchive(gomock.Any(), roleName).Return([]*model.Event{{ID: uuid.New(), Name: "A1", StartsAt: &starts, Status: &status}}, nil)

	res, err := svc.GetEventsByRole(ctx, roleName)
	require.NoError(t, err)
	require.Len(t, res, 1)

	res2, err := svc.GetArchiveEvents(ctx, roleName)
	require.NoError(t, err)
	require.Len(t, res2, 1)
}

// helper to create simple model.Event
func makeModelEvent(id uuid.UUID, name string, starts *time.Time, status *string) *model.Event {
	return &model.Event{ID: id, Name: name, StartsAt: starts, Status: status}
}

// Вспомогательная функция: возвращает указатель на строку.
// Используется для составления DTO в тестах (ptrString("..."))
func ptrString(s string) *string { return &s }
