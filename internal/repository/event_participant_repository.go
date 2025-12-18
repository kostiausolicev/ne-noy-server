package repository

import (
	"errors"
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
	CheckParticipant(participant *model.EventParticipant) error
	Participant(eventID uuid.UUID, userId int64) (bool, error)
	UnParticipant(eventID uuid.UUID, userId int64) (bool, error)
}

func (er *eventParticipantRepository) CheckParticipant(participant *model.EventParticipant) error {
	result := er.db.
		Model(&model.EventParticipant{}).
		Where("event_id = ? AND user_id = ?", participant.EventID, participant.UserID).
		Updates(map[string]interface{}{
			"is_checked":      true,
			"check_timestamp": participant.CheckTimestamp,
			"check_lat":       participant.CheckLat,
			"check_long":      participant.CheckLong,
			"check_type":      participant.CheckType,
			"check_author":    participant.CheckAuthor,
		})
	if result.RowsAffected == 0 {
		return errors.New("participant not exist")
	}
	return result.Error
}

// Participant TODO сделать сырой запрос с подзапросом
func (er *eventParticipantRepository) Participant(eventId uuid.UUID, userId int64) (bool, error) {
	user := model.User{}
	er.db.Table("users").Select("id").Where("vk_id = ?", userId).Scan(&user)
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

func (er *eventParticipantRepository) UnParticipant(eventId uuid.UUID, userId int64) (bool, error) {
	sub := er.db.Table(`event_participants`).
		Select("event_participants.id").
		Joins(`INNER JOIN users ON event_participants.user_id = users.id`).
		Where(`event_participants.event_id = ? AND susers.vk_id = ?`, eventId, userId)

	result := er.db.
		Where("id IN (?)", sub).
		Delete(&model.EventParticipant{})
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}
