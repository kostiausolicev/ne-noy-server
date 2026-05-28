package event

import (
	"context"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/repository"
	"ne_noy/internal/service"

	"github.com/google/uuid"
)

func NewEventService(
	r repository.EventBaseRepository,
	u service.UserService,
	rr repository.RoleRepository,
) EventService {
	return eventService{r: r, u: u, rr: rr}
}

type EventService interface {
	PublishEvent(ctx context.Context, eventId uuid.UUID) error
	GetAll(ctx context.Context, vkId int64) ([]dto.EventMiniDto, error)
	GetEventsByRole(ctx context.Context, role string) ([]dto.EventMiniDto, error)
	GetArchiveEvents(ctx context.Context, role string) ([]dto.EventMiniDto, error)
}

type eventService struct {
	r  repository.EventBaseRepository
	rr repository.RoleRepository
	u  service.UserService
}

func (e eventService) PublishEvent(ctx context.Context, eventId uuid.UUID) error {
	return e.r.Publish(ctx, eventId)
}

func (e eventService) GetAll(ctx context.Context, vkId int64) ([]dto.EventMiniDto, error) {
	// Получаем пользователя
	user, err := e.u.GetUserByVkId(ctx, vkId)
	if err != nil {
		return nil, err
	}
	var events []*events.EventView

	if user.Role.Name == config.RoleAdmin {
		// Если админ, возвращаем все мероприятия
		events, err = e.r.GetAll(ctx, nil, nil)
	} else {
		// Иначе, возвращаем только те, где пользователь организатор
		events, err = e.r.GetAllByOrg(ctx, *user.ID)
	}
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(ctx, events)
}

func (e eventService) GetEventsByRole(ctx context.Context, role string) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAll(ctx, &role, nil)
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(ctx, events)
}

func (e eventService) GetArchiveEvents(ctx context.Context, role string) ([]dto.EventMiniDto, error) {
	archived := false
	events, err := e.r.GetAll(ctx, &role, &archived)
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(ctx, events)
}

func (e eventService) parseModelsToDtos(_ context.Context, events []*events.EventView) ([]dto.EventMiniDto, error) {
	eventsDto := make([]dto.EventMiniDto, len(events))

	for i, event := range events {
		// преобразуем организаторов
		orgs := make([]dto.UserMiniDto, len(event.Orgs))
		for j, org := range event.Orgs {
			orgs[j] = userToMiniDto(org)
		}

		// преобразуем участников
		participants := make([]dto.UserMiniDto, len(event.Participants))
		for j, participant := range event.Participants {
			participants[j] = userToMiniDto(participant)
		}

		eventsDto[i] = dto.EventMiniDto{
			ID:                event.ID,
			Title:             event.Name,
			StartsAt:          event.StartsAt,
			ParticipantsCount: event.ParticipantsCount,
			Orgs:              orgs,
			Participants:      participants,
			Status:            event.Status,
			Type:              event.Type,
		}
	}

	return eventsDto, nil
}

func userToMiniDto(user model.User) dto.UserMiniDto {
	return dto.UserMiniDto{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		VkId:      user.VkID,
		PhotoURL:  user.PhotoURL,
	}
}
