package repository

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type eventParticipantRepository struct {
	db *gorm.DB
}

func NewEventParticipantRepository(db *gorm.DB) EventParticipantRepository {
	return &eventParticipantRepository{db: db}
}

type EventParticipantRepository interface {
	Participant(eventId, userId uuid.UUID) (bool, error)
	UnParticipant(eventId, userId uuid.UUID) (bool, error)
}

func (er eventParticipantRepository) Participant(eventId, userId uuid.UUID) (bool, error) {
	eventParticipant := model.EventParticipant{
		EventID: eventId,
		UserID:  userId,
	}
	result := er.db.Create(&eventParticipant)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (er eventParticipantRepository) UnParticipant(eventId, userId uuid.UUID) (bool, error) {
	result := er.db.
		Where(`event_id = ?, user_id = ?`, eventId, userId).
		Delete(&model.EventParticipant{})
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}
