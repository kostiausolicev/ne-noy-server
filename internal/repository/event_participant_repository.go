package repository

import (
	"context"
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

func (er *eventParticipantRepository) withScope(ctx context.Context) *gorm.DB {
	return er.db.WithContext(ctx)
}

type EventParticipantRepository interface {
	CheckParticipant(ctx context.Context, participant *model.EventParticipant) error
	Participant(ctx context.Context, eventID uuid.UUID, userId int64) (bool, error)
	UnParticipant(ctx context.Context, eventID uuid.UUID, userId int64) (bool, error)
}

func (er *eventParticipantRepository) CheckParticipant(ctx context.Context, participant *model.EventParticipant) error {
	result := er.withScope(ctx).
		Model(&model.EventParticipant{}).
		Where("event_id = ? AND user_id = ?", participant.EventID, participant.UserID).
		Updates(map[string]interface{}{
			"is_checked":      participant.IsChecked,
			"check_timestamp": participant.CheckTimestamp,
			"check_lat":       participant.CheckLat,
			"check_lon":       participant.CheckLong,
			"check_type":      participant.CheckType,
			"check_author":    participant.CheckAuthor,
		})
	if result.RowsAffected == 0 {
		return errors.New("participant not exist")
	}
	return result.Error
}

func (er *eventParticipantRepository) Participant(ctx context.Context, eventId uuid.UUID, userId int64) (bool, error) {
	user := model.User{}
	er.withScope(ctx).
		Table("users").
		Select("id").
		Where("vk_id = ?", userId).
		Scan(&user)
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

func (er *eventParticipantRepository) UnParticipant(ctx context.Context, eventId uuid.UUID, userId int64) (bool, error) {
	sub := er.withScope(ctx).
		Table("event_participants").
		Select("event_participants.id").
		Joins("INNER JOIN users ON event_participants.user_id = users.id").
		Where("event_participants.event_id = ? AND users.vk_id = ?", eventId, userId)

	result := er.db.
		Where("id IN (?)", sub).
		Delete(&model.EventParticipant{})
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}
