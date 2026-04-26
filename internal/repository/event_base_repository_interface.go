package repository

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/model/events/as_team"
	"ne_noy/internal/model/events/as_test"

	"github.com/google/uuid"
)

type EventRepository interface {
	// GetAll возвращает все записи мероприятий с фильтрами
	// 	roleCode - код роли, по которой нужно отфильтровать мероприятия
	//  archived - показать архивные мероприятия
	GetAll(ctx context.Context, roleCode *string, archived *bool) ([]*events.EventView, error)

	// GetAllByOrg возвращает все мероприятия, в которых есть переданный организатор
	//  orgId - ID организатора, по которому нужно отфильтровать мероприятия
	GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*events.EventView, error)

	// GetLocationById возвращает данные о местоположении мероприятия
	//  id - ID мероприятия
	//  returns - lat, long широта и долгота мероприятия
	GetLocationById(ctx context.Context, id uuid.UUID) (lat, long *float64, err error)

	// ExistUserParticipationInEvent проверяет, есть ли у пользователя участие в мероприятии
	//  eventId - ID мероприятия
	//  userId - ID пользователя
	ExistUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) (bool, error)

	// GetEventOrgs получить организаторов мероприятия
	//  id - ID мероприятия
	//  limit - ограничение на количество возвращаемых организаторов
	GetEventOrgs(ctx context.Context, id uuid.UUID, limit int) ([]model.User, error)

	// GetByVkPollId получить мероприятие по ссылке на пользователя
	//  pollId - ID опроса ВКонтакте, по которому нужно найти мероприятие
	GetByVkPollId(ctx context.Context, pollId int64) (*as_event.AsEvent, error)

	// GetEventById получить мероприятие с типом мероприятие по id
	//  id - ID мероприятия
	GetEventById(ctx context.Context, id uuid.UUID) (*as_event.AsEvent, error)

	// GetTeamById получить мероприятие с типом командное мероприятие по id
	//  id - ID мероприятия
	GetTeamById(ctx context.Context, id uuid.UUID) (*as_team.AsTeam, error)

	// GetTestById получить мероприятие с типом тест по id
	//  id - ID мероприятия
	GetTestById(ctx context.Context, id uuid.UUID) (*as_test.AsTest, error)

	// GetParticipants получить список участников мероприятия
	//  id - ID мероприятия
	//  limit - ограничение на число возвращаемых участников
	GetParticipants(ctx context.Context, id uuid.UUID, limit int) ([]as_event.EventParticipant, error)

	// CreateEvent создать мероприятие с типом мероприятие
	CreateEvent(ctx context.Context, event *as_event.AsEvent) (*as_event.AsEvent, error)

	// CreateTeam создать мероприятие с типом командное мероприятие
	CreateTeam(ctx context.Context, team *as_team.AsTeam) (*as_team.AsTeam, error)

	// CreateTest создать мероприятие с типом тест
	CreateTest(ctx context.Context, test *as_test.AsTest) (*as_test.AsTest, error)

	// Update updates as_event fields and optionally replaces organizers and available roles.
	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*events.EventView, error)

	// Delete removes an as_event profile by identifier.
	Delete(ctx context.Context, id uuid.UUID, eventType string) error
}
