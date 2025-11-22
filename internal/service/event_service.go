package service

import (
	"ne_noy/internal/dto"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

func NewEventService(r repository.EventRepository) EventService {
	return eventService{r: r}
}

type EventService interface {
	GetEvent(id uuid.UUID) (dto.EventDto, error)
	GetEventsByRole(roleId uuid.UUID) ([]dto.EventMiniDto, error)
	GetArchiveEvents() ([]dto.EventMiniDto, error)
	GetEduEvents() ([]dto.EventMiniDto, error)
	ParticipantToEvent(eventID, userID uuid.UUID) (bool, error)
	UpParticipantToEvent(eventID, userID uuid.UUID) (bool, error)
}

type eventService struct {
	r repository.EventRepository
}

func (e eventService) GetEvent(id uuid.UUID) (dto.EventDto, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) GetEventsByRole(roleId uuid.UUID) ([]dto.EventMiniDto, error) {
	events, err := e.r.GetAllByRole(roleId)
	if err != nil {
		return nil, err
	}

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

func (e eventService) GetArchiveEvents() ([]dto.EventMiniDto, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) GetEduEvents() ([]dto.EventMiniDto, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) ParticipantToEvent(eventID, userID uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventService) UpParticipantToEvent(eventID, userID uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}
