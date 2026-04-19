package service

import (
	"context"
	"fmt"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"
	"sync"

	"github.com/google/uuid"
)

func NewEventService(
	r repository.EventRepository,
	u UserService,
	rr repository.RoleRepository,
) EventService {
	return eventService{r: r, u: u, rr: rr}
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
	r  repository.EventRepository
	rr repository.RoleRepository
	u  UserService
}

func (e eventService) CreateEvent(ctx context.Context, eventDto dto.CreateUpdateEventDto) (*dto.EventMiniDto, error) {
	eventId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	event, err := e.parseDtoToModel(ctx, eventDto, eventId)
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
	if eventDto.AdAddress != nil {
		fields["additional_address"] = eventDto.AdAddress
	}
	if eventDto.VkPollAnswerID != nil {
		fields["vk_poll_answer_id"] = eventDto.VkPollAnswerID
	}
	if eventDto.VkVoteID != nil {
		fields["vk_vote_id"] = eventDto.VkVoteID
	}
	if eventDto.Lat != nil {
		fields["lat"] = eventDto.Lat
	}
	if eventDto.Long != nil {
		fields["lon"] = eventDto.Long
	}
	if eventDto.StartsAt != nil {
		fields["starts_at"] = eventDto.StartsAt
	}
	if eventDto.EndsAt != nil {
		fields["ends_at"] = eventDto.EndsAt
	}
	if eventDto.Status != nil {
		fields["status"] = eventDto.Status
	}

	// Формируем слайс организаторов и ролей (если переданы)
	orgs := make([]model.User, 0)
	for _, org := range eventDto.Orgs {
		orgs = append(orgs, model.User{ID: org.ID})
	}
	roles := make([]model.Role, 0)
	for _, roleCode := range eventDto.AvailableRoles {
		role, err := e.rr.GetByCode(ctx, roleCode)
		if err != nil {
			return nil, err
		}
		roles = append(roles, *role)
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
		events, err = e.r.GetAllByOrg(ctx, *user.ID)
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
	eventType, err := e.r.GetEventTypeById(ctx, id)
	if err != nil {
		return nil, err
	}

	switch eventType {
	case model.EventTypeEvent:
		event, err := e.r.GetEventById(ctx, id)
		if err != nil {
			return nil, err
		}
		return e.parseEventModelToDto(ctx, event, userId)
	case model.EventTypeActivity:
		event, err := e.r.GetActivityById(ctx, id)
		if err != nil {
			return nil, err
		}
		return e.parseActivityModelToDto(ctx, event, userId)
	case model.EventTypeTeam:
		event, err := e.r.GetTeamById(ctx, id)
		if err != nil {
			return nil, err
		}
		return e.parseTeamModelToDto(ctx, event, userId)
	case model.EventTypePoll:
		event, err := e.r.GetPollById(ctx, id)
		if err != nil {
			return nil, err
		}
		return e.parsePollModelToDto(ctx, event, userId)
	case model.EventTypeTest:
		event, err := e.r.GetTestById(ctx, id)
		if err != nil {
			return nil, err
		}
		return e.parseTestModelToDto(ctx, event, userId)
	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}

func (e eventService) buildUsersAndAttachments(relations model.EventRelations) ([]dto.UserMiniDto, []dto.UserMiniDto, []dto.AttachmentDto) {
	var wg = sync.WaitGroup{}
	orgs := make([]dto.UserMiniDto, len(relations.Orgs))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j, org := range relations.Orgs {
			orgs[j] = dto.UserMiniDto{
				ID:        org.ID,
				FirstName: org.FirstName,
				LastName:  org.LastName,
				VkId:      org.VkID,
				PhotoURL:  org.PhotoURL,
			}
		}
	}()

	participants := make([]dto.UserMiniDto, len(relations.EventParticipants))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j, ep := range relations.EventParticipants {
			participants[j] = dto.UserMiniDto{
				ID:        ep.User.ID,
				FirstName: ep.User.FirstName,
				LastName:  ep.User.LastName,
				VkId:      ep.User.VkID,
				PhotoURL:  ep.User.PhotoURL,
			}
		}
	}()

	attachments := make([]dto.AttachmentDto, len(relations.Attachments))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i, att := range relations.Attachments {
			attachments[i] = dto.AttachmentDto{
				ID:    att.Attachment.ID,
				Url:   att.Attachment.Url,
				Title: att.Attachment.Filename,
			}
		}
	}()

	wg.Wait()
	return orgs, participants, attachments
}

func (e eventService) parseEventModelToDto(ctx context.Context, event *model.EventAsEvent, userId int64) (*dto.EventDto, error) {
	orgs, participants, attachments := e.buildUsersAndAttachments(event.EventRelations)
	isParticipant, err := e.r.GetUserParticipationInEvent(ctx, event.ID, userId)
	if err != nil {
		return nil, err
	}

	eventDto := &dto.EventDto{
		ID:                       event.ID,
		VkPostId:                 event.VkPostID,
		PhotoURL:                 event.Cover,
		VkPollAnswerID:           event.VkPollAnswerID,
		VkVoteID:                 event.VkVoteID,
		Lat:                      event.Lat,
		Long:                     event.Lon,
		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        event.ParticipantsCount,
		Orgs:                     orgs,
		Address:                  event.Address,
		AdAddress:                event.AdditionalAddress,
		Participants:             participants,
		StartsAt:                 event.StartsAt,
		EndsAt:                   *event.EndsAt,
		Status:                   event.Status,
		CurrentUserIsParticipant: &isParticipant,
	}
	return eventDto, nil
}

func (e eventService) parseActivityModelToDto(ctx context.Context, event *model.EventAsActivity, userId int64) (*dto.EventDto, error) {
	orgs, participants, attachments := e.buildUsersAndAttachments(event.EventRelations)
	isParticipant, err := e.r.GetUserParticipationInEvent(ctx, event.ID, userId)
	if err != nil {
		return nil, err
	}

	return &dto.EventDto{
		ID:                       event.ID,
		PhotoURL:                 event.Cover,
		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        event.ParticipantsCount,
		Orgs:                     orgs,
		Participants:             participants,
		StartsAt:                 event.StartsAt,
		EndsAt:                   *event.EndsAt,
		Status:                   event.Status,
		CurrentUserIsParticipant: &isParticipant,
	}, nil
}

func (e eventService) parseTeamModelToDto(ctx context.Context, event *model.EventAsTeam, userId int64) (*dto.EventDto, error) {
	orgs, participants, attachments := e.buildUsersAndAttachments(event.EventRelations)
	isParticipant, err := e.r.GetUserParticipationInEvent(ctx, event.ID, userId)
	if err != nil {
		return nil, err
	}

	return &dto.EventDto{
		ID:                       event.ID,
		VkPostId:                 event.VkPostID,
		PhotoURL:                 event.Cover,
		Lat:                      event.Lat,
		Long:                     event.Lon,
		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        event.ParticipantsCount,
		Orgs:                     orgs,
		Address:                  event.Address,
		AdAddress:                event.AdditionalAddress,
		Participants:             participants,
		StartsAt:                 event.StartsAt,
		EndsAt:                   *event.EndsAt,
		Status:                   event.Status,
		CurrentUserIsParticipant: &isParticipant,
	}, nil
}

func (e eventService) parsePollModelToDto(ctx context.Context, event *model.EventAsPoll, userId int64) (*dto.EventDto, error) {
	orgs, participants, attachments := e.buildUsersAndAttachments(event.EventRelations)
	isParticipant, err := e.r.GetUserParticipationInEvent(ctx, event.ID, userId)
	if err != nil {
		return nil, err
	}

	return &dto.EventDto{
		ID:                       event.ID,
		VkPostId:                 event.VkPostID,
		PhotoURL:                 event.Cover,
		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        event.ParticipantsCount,
		Orgs:                     orgs,
		Participants:             participants,
		StartsAt:                 event.StartsAt,
		EndsAt:                   *event.EndsAt,
		Status:                   event.Status,
		CurrentUserIsParticipant: &isParticipant,
	}, nil
}

func (e eventService) parseTestModelToDto(ctx context.Context, event *model.EventAsTest, userId int64) (*dto.EventDto, error) {
	orgs, participants, attachments := e.buildUsersAndAttachments(event.EventRelations)
	isParticipant, err := e.r.GetUserParticipationInEvent(ctx, event.ID, userId)
	if err != nil {
		return nil, err
	}

	return &dto.EventDto{
		ID:                       event.ID,
		VkPostId:                 event.VkPostID,
		PhotoURL:                 event.Cover,
		Title:                    event.Name,
		Description:              event.Description,
		Attachments:              attachments,
		ParticipantsCount:        event.ParticipantsCount,
		Orgs:                     orgs,
		Participants:             participants,
		StartsAt:                 event.StartsAt,
		EndsAt:                   *event.EndsAt,
		Status:                   event.Status,
		CurrentUserIsParticipant: &isParticipant,
	}, nil
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

func (e eventService) parseDtoToModel(ctx context.Context, eventDto dto.CreateUpdateEventDto, eventId uuid.UUID) (*model.Event, error) {
	// Создаем event с минимальными данными
	event := model.Event{
		ID:        eventId,
		EventType: model.EventTypeEvent,
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
	if eventDto.AdAddress != nil {
		event.AdditionalAddress = eventDto.AdAddress
	}
	if eventDto.VkPollAnswerID != nil {
		event.VkPollAnswerID = eventDto.VkPollAnswerID
	}
	if eventDto.VkVoteID != nil {
		event.VkVoteID = eventDto.VkVoteID
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
	if eventDto.EndsAt != nil {
		event.EndsAt = eventDto.EndsAt
	}

	// Создаем организаторов только с ID
	event.Orgs = make([]model.User, len(eventDto.Orgs))
	for i, org := range eventDto.Orgs {
		event.Orgs[i] = model.User{ID: org.ID}
	}

	// Создаем роли только с ID (теперь AvailableRoles - это []uuid.UUID)
	event.AvailableRoles = make([]model.Role, len(eventDto.AvailableRoles))
	for i, roleCode := range eventDto.AvailableRoles {
		role, err := e.rr.GetByCode(ctx, roleCode)
		if err != nil {
			return nil, err
		}
		event.AvailableRoles[i] = *role
	}

	return &event, nil
}
