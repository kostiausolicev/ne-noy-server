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
	Participant(eventID uuid.UUID, userId int64) (bool, error)
	UnParticipant(eventID uuid.UUID, userId int64) (bool, error)
}

func (er eventParticipantRepository) Participant(eventId uuid.UUID, userId int64) (bool, error) {
	user := model.User{}
	er.db.Table("user").Select("id").Where("vk_id = ?", userId).Scan(&user)
	eventParticipant := model.EventParticipant{
		EventID: eventId,
		UserID:  user.ID,
	}
	result := er.db.Create(&eventParticipant)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (er eventParticipantRepository) UnParticipant(eventId uuid.UUID, userId int64) (bool, error) {
	sub := er.db.Table(`event_participant`).
		Select("event_participant.id").
		Joins(`INNER JOIN "user" ON event_participant.user_id = "user".id`).
		Where(`event_participant.event_id = ? AND "user".vk_id = ?`, eventId, userId)

	result := er.db.
		Where("id IN (?)", sub).
		Delete(&model.EventParticipant{})
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}
