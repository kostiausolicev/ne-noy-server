package service

import (
	"ne_noy/internal/repository"

	"github.com/google/uuid"
)

type eventParticipantService struct {
	r repository.EventParticipantRepository
}

func NewEventParticipantService(r repository.EventParticipantRepository) EventParticipantService {
	return eventParticipantService{r: r}
}

type EventParticipantService interface {
	ParticipantToEvent(eventID, userID uuid.UUID) (bool, error)
	UpParticipantToEvent(eventID, userID uuid.UUID) (bool, error)
}

func (eps eventParticipantService) ParticipantToEvent(eventID, userID uuid.UUID) (bool, error) {
	return eps.r.Participant(eventID, userID)
}

func (eps eventParticipantService) UpParticipantToEvent(eventID, userID uuid.UUID) (bool, error) {
	return eps.r.UnParticipant(eventID, userID)
}
