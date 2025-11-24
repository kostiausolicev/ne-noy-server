package service

import (
	"ne_noy/internal/dto"
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
	ParticipantToEvent(eventID, userID uuid.UUID) (bool, error)
	UpParticipantToEvent(eventID, userID uuid.UUID) (bool, error)
}

func (eps eventParticipantService) ParticipantToEvent(eventID, userID uuid.UUID) (bool, error) {
	return eps.epr.Participant(eventID, userID)
}

func (eps eventParticipantService) UpParticipantToEvent(eventID, userID uuid.UUID) (bool, error) {
	return eps.epr.UnParticipant(eventID, userID)
}

func (eps eventParticipantService) CheckParticipant(participantData dto.EventParticipantDto) error {
	eventId := participantData.EventID
	location, err := eps.er.GetEventLocationData(eventId)
	if err != nil {
		return err
	}
	_, err = eps.er.Update(location)
	if err != nil {
		return err
	}
	return nil
}
