package service

import (
	"errors"
	"ne_noy/internal/dto"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

type eventParticipantService struct {
	epr repository.EventParticipantRepository
	er  repository.EventRepository
}

func NewEventParticipantService(epr repository.EventParticipantRepository, er repository.EventRepository) EventParticipantService {
	return eventParticipantService{epr: epr, er: er}
}

type EventParticipantService interface {
	ParticipantToEvent(eventID uuid.UUID, userVkID int64) (bool, error)
	UpParticipantToEvent(eventID uuid.UUID, userVkID int64) (bool, error)
	CheckParticipant(participantData dto.CheckEventParticipant) error
}

func (eps eventParticipantService) ParticipantToEvent(eventID uuid.UUID, userVkID int64) (bool, error) {
	return eps.epr.Participant(eventID, userVkID)
}

func (eps eventParticipantService) UpParticipantToEvent(eventID uuid.UUID, userVkID int64) (bool, error) {
	return eps.epr.UnParticipant(eventID, userVkID)
}

func (eps eventParticipantService) CheckParticipant(participantData dto.CheckEventParticipant) (err error) {
	switch participantData.CheckType {
	case "Personal QR":
		{
			err = eps.checkByAdmin(participantData)
		}
	case "Admin panel":
		{
			err = eps.checkByAdmin(participantData)
		}
	case "Event QR":
		{

		}
	}
	return err
}

func (eps eventParticipantService) checkByAdmin(participantData dto.CheckEventParticipant) error {
	orgs, err := eps.er.GetEventOrgs(participantData.EventId)
	if err != nil {
		return err
	}
	for _, eventOrg := range orgs {
		if eventOrg.VkID == *participantData.CheckAuthorVkId {
			participant := model.EventParticipant{
				EventID:        participantData.EventId,
				UserID:         participantData.UserId,
				IsChecked:      true,
				CheckTimestamp: &participantData.Timestamp,
				CheckLat:       participantData.Lat,
				CheckLong:      participantData.Long,
				CheckType:      participantData.CheckType,
				CheckAuthor:    &eventOrg.ID,
			}
			err := eps.epr.CheckParticipant(&participant)
			return err
		}
	}
	return errors.New("participant not exist")
}
