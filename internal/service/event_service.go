package service

import (
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
	CreateEvent(eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error)
	UpdateEvent(eventId uuid.UUID, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error)
	GetAll(vkId int64) ([]dto.EventMiniDto, error)
	GetEventParticipants(id uuid.UUID) ([]dto.EventParticipantDto, error)
	GetEvent(id uuid.UUID, userVkId int64) (*dto.EventDto, error)
	GetEventsByRole(roleId uuid.UUID) ([]dto.EventMiniDto, error)
	GetArchiveEvents(roleId uuid.UUID) ([]dto.EventMiniDto, error)
	GetEduEvents() ([]dto.EventMiniDto, error)
}

type eventService struct {
	r repository.EventRepository
	u UserService
}

func (e eventService) CreateEvent(eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error) {
	eventId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	event, err := e.parseDtoToModel(eventDto, eventId)
	if err != nil {
		return nil, err
	}

	newEvent, err := e.r.Create(event)
	if err != nil {
		return nil, err
	}
	return e.parseModelToDto(newEvent)
}

func (e eventService) UpdateEvent(eventId uuid.UUID, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error) {
	event, err := e.parseDtoToModel(eventDto, eventId)
	if err != nil {
		return nil, err
	}
	updatedEvent, err := e.r.Update(event)
	if err != nil {
		return nil, err
	}

	return e.parseModelToDto(updatedEvent)
}

func (e eventService) GetAll(vkId int64) ([]dto.EventMiniDto, error) {
	// Получаем пользователя
	user, err := e.u.GetUserByVkId(vkId)
	if err != nil {
		return nil, err
	}
	var events []*model.Event

	if user.Role.Name == config.RoleAdmin {
		// Если админ, возвращаем все мероприятия
		events, err = e.r.GetAll()
	} else {
		// Иначе, возвращаем только те, где пользователь организатор
		events, err = e.r.GetAllByOrg(user.ID)
	}
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(events)
}

func (e eventService) GetEventParticipants(id uuid.UUID) ([]dto.EventParticipantDto, error) {
	eventParticipants, err := e.r.GetParticipants(id)
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

func (e eventService) GetEvent(id uuid.UUID, userId int64) (*dto.EventDto, error) {
	event, err := e.r.GetById(id)
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
	count, err := e.r.CountParticipants(event.ID)
	if err != nil {
		return nil, err
	}
	attachments := make([]string, len(event.Attachments))
	for i, att := range event.Attachments {
		attachments[i] = att.AttachmentLink
	}
	isParticipant, err := e.r.GetUserParticipationInEvent(event.ID, userId)
	if err != nil {
		return nil, err
	}

	eventDto := &dto.EventDto{
		ID:         event.ID,
		VkPostLink: event.VkPostLink,
		PhotoURL:   event.Cover,

		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        count,
		Orgs:                     orgs,
		Address:                  event.Address,
		Participants:             participants,
		StartsAt:                 event.StartsAt,
		CurrentUserIsParticipant: &isParticipant,
	}
	return eventDto, nil
}

func (e eventService) GetEventsByRole(roleId uuid.UUID) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAllByRole(roleId)
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(events)
}

func (e eventService) GetArchiveEvents(roleId uuid.UUID) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAllArchive(roleId)
	if err != nil {
		return nil, err
	}

	return e.parseModelsToDtos(events)
}

func (e eventService) GetEduEvents() ([]dto.EventMiniDto, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) parseModelsToDtos(events []*model.Event) ([]dto.EventMiniDto, error) {
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
		count, err := e.r.CountParticipants(event.ID)
		if err != nil {
			return nil, err
		}

		eventsDto[i] = dto.EventMiniDto{
			ID:                event.ID,
			Title:             event.Name,
			StartsAt:          *event.StartsAt,
			ParticipantsCount: count,
			Orgs:              orgs,
			Participants:      participants,
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
	count, err := e.r.CountParticipants(event.ID)
	if err != nil {
		return nil, err
	}

	eventDto := &dto.EventMiniDto{
		ID:                event.ID,
		Title:             event.Name,
		StartsAt:          *event.StartsAt,
		ParticipantsCount: count,
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
	if eventDto.VkPostLink != nil {
		event.VkPostLink = eventDto.VkPostLink
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
	if eventDto.Status != "" {
		event.Status = &eventDto.Status
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
