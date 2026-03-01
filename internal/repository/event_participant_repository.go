package repository

import (
	"context"
	"ne_noy/internal/apperror"
	"ne_noy/internal/model"
	"slices"

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
	Participant(ctx context.Context, eventID uuid.UUID, userVkId int64, prepareType string) (bool, error)
	ParticipantById(ctx context.Context, eventID uuid.UUID, userId uuid.UUID, prepareType string) (bool, error)
	UnParticipant(ctx context.Context, eventID uuid.UUID, userId int64) (bool, error)
}

func (er *eventParticipantRepository) CheckParticipant(ctx context.Context, participant *model.EventParticipant) error {
	var exists bool
	result := er.withScope(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM event_participants WHERE event_id = ? AND user_id = ?)", participant.EventID, participant.UserID).
		Scan(&exists)

	if result.Error != nil {
		return result.Error
	}

	if !exists {
		return apperror.ParticipantNotExistErr
	}

	result = er.withScope(ctx).
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
	return result.Error
}

func (er *eventParticipantRepository) Participant(ctx context.Context, eventId uuid.UUID, userVkId int64, prepareType string) (bool, error) {
	user := model.User{}
	event := model.Event{}
	er.withScope(ctx).
		Table(user.TableName()).
		Select("\"users\".id").
		Joins("Role").
		Where("\"users\".vk_id = ?", userVkId).
		First(&user)
	er.withScope(ctx).
		Table(event.TableName()).
		Select("id").
		Preload("AvailableRoles").
		Where("id = ?", eventId).
		First(&event)
	if !slices.Contains(event.AvailableRoles, *user.Role) {
		return false, apperror.UserRoleNotInAvailableRolesErr
	}
	eventParticipant := model.EventParticipant{
		EventID:     eventId,
		UserID:      user.ID,
		PrepareType: prepareType,
	}
	result := er.db.WithContext(ctx).Create(&eventParticipant)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

func (er *eventParticipantRepository) ParticipantById(ctx context.Context, eventId uuid.UUID, userId uuid.UUID, prepareType string) (bool, error) {
	user := model.User{}
	event := model.Event{}
	er.withScope(ctx).
		Table(user.TableName()).
		Select("\"users\".id").
		Joins("Role").
		Where("\"users\".id = ?", userId).
		First(&user)
	er.withScope(ctx).
		Table(event.TableName()).
		Select("id").
		Preload("AvailableRoles").
		Where("id = ?", eventId).
		First(&event)
	if !slices.Contains(event.AvailableRoles, *user.Role) {
		return false, apperror.UserRoleNotInAvailableRolesErr
	}
	eventParticipant := model.EventParticipant{
		EventID:     eventId,
		UserID:      userId,
		PrepareType: prepareType,
	}
	result := er.db.WithContext(ctx).Create(&eventParticipant)
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
