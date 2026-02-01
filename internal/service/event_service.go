package service

import (
	"context"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

func NewEventService(r repository.EventRepository, u UserService) EventService {
	return eventService{r: r, u: u}
}

type EventService interface {
	CreateEvent(ctx context.Context, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error)
	PublishEvent(ctx context.Context, eventId uuid.UUID) (*dto.EventMiniDto, error)
	UpdateEvent(ctx context.Context, eventId uuid.UUID, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error)
	GetAll(ctx context.Context, vkId int64) ([]dto.EventMiniDto, error)
	GetEventParticipants(ctx context.Context, id uuid.UUID) ([]dto.EventParticipantDto, error)
	GetEvent(ctx context.Context, id uuid.UUID, userVkId int64) (*dto.EventDto, error)
	GetEventsByRole(ctx context.Context, role string) ([]dto.EventMiniDto, error)
	GetArchiveEvents(ctx context.Context, role string) ([]dto.EventMiniDto, error)
	GetEduEvents(ctx context.Context) ([]dto.EventMiniDto, error)
}

type eventService struct {
	r repository.EventRepository
	u UserService
}

func (e eventService) CreateEvent(ctx context.Context, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error) {
	eventId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	event, err := e.parseDtoToModel(eventDto, eventId)
	if err != nil {
		return nil, err
	}

	newEvent, err := e.r.Create(ctx, event)
	if err != nil {
		return nil, err
	}
	return e.parseModelToDto(newEvent)
}

func (e eventService) PublishEvent(ctx context.Context, eventId uuid.UUID) (*dto.EventMiniDto, error) {
	fields := make(map[string]interface{})
	fields["status"] = "ACTIVE"
	updatedEvent, err := e.r.Update(ctx, eventId, fields, nil, nil)
	if err != nil {
		return nil, err
	}

	return e.parseModelToDto(updatedEvent)
}

func (e eventService) UpdateEvent(ctx context.Context, eventId uuid.UUID, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error) {
	// Собираем поля которые нужно обновить
	fields := make(map[string]interface{})
	if eventDto.Title != nil {
		fields["name"] = *eventDto.Title
	}
	if eventDto.Description != nil {
		fields["description"] = eventDto.Description
	}
	if eventDto.VkPostId != nil {
		fields["vk_post_id"] = eventDto.VkPostId
	}
	if eventDto.PhotoURL != nil {
		fields["cover"] = eventDto.PhotoURL
	}
	if eventDto.Address != nil {
		fields["address"] = eventDto.Address
	}
	if eventDto.Lat != nil {
		fields["lat"] = eventDto.Lat
	}
	if eventDto.Long != nil {
		fields["long"] = eventDto.Long
	}
	if eventDto.StartsAt != nil {
		fields["starts_at"] = eventDto.StartsAt
	}

	// Формируем слайс организаторов и ролей (если переданы)
	orgs := make([]model.User, 0)
	for _, orgId := range eventDto.Orgs {
		orgs = append(orgs, model.User{ID: orgId})
	}
	roles := make([]model.Role, 0)
	for _, roleId := range eventDto.AvailableRoles {
		roles = append(roles, model.Role{ID: roleId})
	}

	updatedEvent, err := e.r.Update(ctx, eventId, fields, orgs, roles)
	if err != nil {
		return nil, err
	}

	return e.parseModelToDto(updatedEvent)
}

func (e eventService) GetAll(ctx context.Context, vkId int64) ([]dto.EventMiniDto, error) {
	// Получаем пользователя
	user, err := e.u.GetUserByVkId(ctx, vkId)
	if err != nil {
		return nil, err
	}
	var events []*model.Event

	if user.Role.Name == config.RoleAdmin {
		// Если админ, возвращаем все мероприятия
		events, err = e.r.GetAll(ctx)
	} else {
		// Иначе, возвращаем только те, где пользователь организатор
		events, err = e.r.GetAllByOrg(ctx, user.ID)
	}
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(ctx, events)
}

func (e eventService) GetEventParticipants(ctx context.Context, id uuid.UUID) ([]dto.EventParticipantDto, error) {
	eventParticipants, err := e.r.GetParticipants(ctx, id)
	if err != nil {
		return nil, err
	}

	participants := make([]dto.EventParticipantDto, len(eventParticipants))
	for i, ep := range eventParticipants {
		userDto := dto.UserMiniDto{
			ID:        ep.User.ID,
			FirstName: ep.User.FirstName,
			LastName:  ep.User.LastName,
			VkId:      ep.User.VkID,
			PhotoURL:  ep.User.PhotoURL,
		}
		participants[i] = dto.EventParticipantDto{
			User:           userDto,
			IsChecked:      ep.IsChecked,
			CheckTimestamp: ep.CheckTimestamp,
		}
	}
	return participants, nil
}

func (e eventService) GetEvent(ctx context.Context, id uuid.UUID, userId int64) (*dto.EventDto, error) {
	event, err := e.r.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	// преобразуем организаторов
	orgs := make([]dto.UserMiniDto, len(event.Orgs))
	for j, org := range event.Orgs {
		orgs[j] = dto.UserMiniDto{
			ID:        org.ID,
			FirstName: org.FirstName,
			LastName:  org.LastName,
			VkId:      org.VkID,
			PhotoURL:  org.PhotoURL,
		}
	}
	// преобразуем участников
	participants := make([]dto.UserMiniDto, len(event.EventParticipants))
	for j, ep := range event.EventParticipants {
		participants[j] = dto.UserMiniDto{
			ID:        ep.User.ID,
			FirstName: ep.User.FirstName,
			LastName:  ep.User.LastName,
			VkId:      ep.User.VkID,
			PhotoURL:  ep.User.PhotoURL,
		}
	}

	attachments := make([]string, len(event.Attachments))
	for i, att := range event.Attachments {
		attachments[i] = att.AttachmentLink
	}
	isParticipant, err := e.r.GetUserParticipationInEvent(ctx, event.ID, userId)
	if err != nil {
		return nil, err
	}

	eventDto := &dto.EventDto{
		ID:       event.ID,
		VkPostId: event.VkPostId,
		PhotoURL: event.Cover,

		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        event.ParticipantsCount,
		Orgs:                     orgs,
		Address:                  event.Address,
		Participants:             participants,
		StartsAt:                 *event.StartsAt,
		Status:                   *event.Status,
		CurrentUserIsParticipant: &isParticipant,
	}
	return eventDto, nil
}

func (e eventService) GetEventsByRole(ctx context.Context, role string) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAllByRole(ctx, role)
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(ctx, events)
}

func (e eventService) GetArchiveEvents(ctx context.Context, role string) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAllArchive(ctx, role)
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(ctx, events)
}

func (e eventService) GetEduEvents(ctx context.Context) ([]dto.EventMiniDto, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) parseModelsToDtos(ctx context.Context, events []*model.Event) ([]dto.EventMiniDto, error) {
	eventsDto := make([]dto.EventMiniDto, len(events))

	for i, event := range events {
		// преобразуем организаторов
		orgs := make([]dto.UserMiniDto, len(event.Orgs))
		for j, org := range event.Orgs {
			orgs[j] = dto.UserMiniDto{
				ID:        org.ID,
				FirstName: org.FirstName,
				LastName:  org.LastName,
				VkId:      org.VkID,
				PhotoURL:  org.PhotoURL,
			}
		}

		// преобразуем участников
		participants := make([]dto.UserMiniDto, len(event.EventParticipants))
		for j, ep := range event.EventParticipants {
			participants[j] = dto.UserMiniDto{
				ID:        ep.User.ID,
				FirstName: ep.User.FirstName,
				LastName:  ep.User.LastName,
				VkId:      ep.User.VkID,
				PhotoURL:  ep.User.PhotoURL,
			}
		}

		eventsDto[i] = dto.EventMiniDto{
			ID:                event.ID,
			Title:             event.Name,
			StartsAt:          *event.StartsAt,
			ParticipantsCount: event.ParticipantsCount,
			Orgs:              orgs,
			Participants:      participants,
			Status:            *event.Status,
		}
	}

	return eventsDto, nil
}

func (e eventService) parseModelToDto(event *model.Event) (*dto.EventMiniDto, error) {
	// преобразуем организаторов
	orgs := make([]dto.UserMiniDto, len(event.Orgs))
	for j, org := range event.Orgs {
		orgs[j] = dto.UserMiniDto{
			ID:        org.ID,
			FirstName: org.FirstName,
			LastName:  org.LastName,
			VkId:      org.VkID,
			PhotoURL:  org.PhotoURL,
		}
	}

	// преобразуем участников
	participants := make([]dto.UserMiniDto, len(event.EventParticipants))
	for j, ep := range event.EventParticipants {
		participants[j] = dto.UserMiniDto{
			ID:        ep.User.ID,
			FirstName: ep.User.FirstName,
			LastName:  ep.User.LastName,
			VkId:      ep.User.VkID,
			PhotoURL:  ep.User.PhotoURL,
		}
	}

	eventDto := &dto.EventMiniDto{
		ID:                event.ID,
		Title:             event.Name,
		StartsAt:          *event.StartsAt,
		ParticipantsCount: event.ParticipantsCount,
		Orgs:              orgs,
		Participants:      participants,
	}

	return eventDto, nil
}

func (e eventService) parseDtoToModel(eventDto dto.CreateUpdateEventDto, eventId uuid.UUID) (*model.Event, error) {
	// Создаем event с минимальными данными
	event := model.Event{
		ID: eventId,
	}

	// Заполняем поля (проверяем на nil)
	if eventDto.Title != nil {
		event.Name = *eventDto.Title
	}
	if eventDto.Description != nil {
		event.Description = eventDto.Description
	}
	if eventDto.VkPostId != nil {
		event.VkPostId = eventDto.VkPostId
	}
	if eventDto.PhotoURL != nil {
		event.Cover = eventDto.PhotoURL
	}
	if eventDto.Address != nil {
		event.Address = eventDto.Address
	}
	if eventDto.Lat != nil {
		event.Lat = eventDto.Lat
	}
	if eventDto.Long != nil {
		event.Long = eventDto.Long
	}
	if eventDto.Status != nil {
		event.Status = eventDto.Status
	}
	if eventDto.StartsAt != nil {
		event.StartsAt = eventDto.StartsAt
	}

	// Создаем организаторов только с ID
	event.Orgs = make([]model.User, len(eventDto.Orgs))
	for i, orgId := range eventDto.Orgs {
		event.Orgs[i] = model.User{ID: orgId}
	}

	// Создаем роли только с ID (теперь AvailableRoles - это []uuid.UUID)
	event.AvailableRoles = make([]model.Role, len(eventDto.AvailableRoles))
	for i, roleId := range eventDto.AvailableRoles {
		event.AvailableRoles[i] = model.Role{ID: roleId}
	}

	return &event, nil
}
