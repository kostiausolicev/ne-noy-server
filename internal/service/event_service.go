package service

import (
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

func NewEventService(r repository.EventRepository) EventService {
	return eventService{r: r}
}

type EventService interface {
	GetEvent(id uuid.UUID, userVkId int64) (*dto.EventDto, error)
	GetEventsByRole(roleId uuid.UUID) ([]dto.EventMiniDto, error)
	GetArchiveEvents(roleId uuid.UUID) ([]dto.EventMiniDto, error)
	GetEduEvents() ([]dto.EventMiniDto, error)
}

type eventService struct {
	r repository.EventRepository
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
		VkPostLink: event.VkPostID,
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

	return e.parseModelToDto(events)
}

func (e eventService) GetArchiveEvents(roleId uuid.UUID) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAllArchive(roleId)
	if err != nil {
		return nil, err
	}

	return e.parseModelToDto(events)
}

func (e eventService) GetEduEvents() ([]dto.EventMiniDto, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) parseModelToDto(events []*model.Event) ([]dto.EventMiniDto, error) {
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
