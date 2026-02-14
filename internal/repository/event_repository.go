package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ne_noy/internal/model"
)

const (
	selectUserFields = `id, vk_id, first_name, last_name, photo_url`
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

type EventRepository interface {
	GetAll(ctx context.Context) ([]*model.Event, error)
	GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*model.Event, error)
	GetAllByRole(ctx context.Context, role string) ([]*model.Event, error)
	GetAllArchive(ctx context.Context, role string) ([]*model.Event, error)
	GetEventLocationData(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userVkId int64) (bool, error)
	GetEventOrgs(ctx context.Context, eventId uuid.UUID) ([]model.User, error)
	GetByVkPollId(ctx context.Context, pollId int64) (*model.Event, error)
	GetById(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetParticipants(ctx context.Context, id uuid.UUID) ([]model.EventParticipant, error)
	Create(ctx context.Context, event *model.Event) (*model.Event, error)
	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*model.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

func (r *eventRepository) GetAll(ctx context.Context) ([]*model.Event, error) {
	var events []*model.Event
	result := r.baseEventQuery(ctx).Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (r *eventRepository) GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event
	result := r.baseEventQuery(ctx).
		Joins("LEFT JOIN event_orgs eo ON eo.event_id = events.id").
		Where("eo.user_id = ?", orgId).
		Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (r *eventRepository) GetUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userVkId int64) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Table("event_participants ep").
		Joins("INNER JOIN users u ON ep.user_id = u.id").
		Where("ep.event_id = ? AND u.vk_id = ?", eventId, userVkId).
		Select("EXISTS (SELECT 1)").
		Scan(&exists).
		Error
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *eventRepository) GetByVkPollId(ctx context.Context, vkPollId int64) (*model.Event, error) {
	var event model.Event
	err := r.db.WithContext(ctx).
		Table("events").
		Select("vk_poll_answer_id").
		Where("vk_vote_id = ?", vkPollId).
		First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) GetEventOrgs(ctx context.Context, eventId uuid.UUID) ([]model.User, error) {
	var orgs []model.User
	err := r.db.WithContext(ctx).
		Table("event_orgs eo").
		Select(selectUserFields).
		Joins("JOIN users u ON u.id = eo.user_id").
		Where("eo.event_id = ?", eventId).
		Scan(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *eventRepository) GetAllByRole(ctx context.Context, role string) ([]*model.Event, error) {
	var events []*model.Event
	result := r.baseEventQuery(ctx).
		Joins("JOIN event_roles er ON er.event_id = events.id").
		Joins("JOIN roles r ON r.id = er.role_id").
		Where("r.name = ? AND events.starts_at > NOW()", role).
		Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (r *eventRepository) GetAllArchive(ctx context.Context, role string) ([]*model.Event, error) {
	var events []*model.Event
	result := r.baseEventQuery(ctx).
		Joins("JOIN event_roles er ON er.event_id = events.id").
		Joins("JOIN roles r ON r.id = er.role_id").
		Where("r.name = ? AND events.starts_at < NOW()", role).
		Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

func (r *eventRepository) GetParticipants(ctx context.Context, id uuid.UUID) ([]model.EventParticipant, error) {
	var participants []model.EventParticipant
	result := r.db.WithContext(ctx).
		Table("event_participants").
		Select(`event_participants.user_id, event_participants.check_timestamp, event_participants.is_checked`).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Where("event_participants.event_id = ?", id).
		Find(&participants)
	if result.Error != nil {
		return nil, result.Error
	}
	return participants, nil
}

func (r *eventRepository) GetById(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var event model.Event
	result := r.db.WithContext(ctx).
		Table(event.TableName()).
		Preload("Orgs", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.
				Select(`event_participants.id, event_participants.user_id, event_participants.event_id`).
				Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Preload("Attachments.Attachment", func(db *gorm.DB) *gorm.DB {
			return db.Select(`attachments.id, attachments.filename, attachments.url`)
		}).
		Select("id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, status, starts_at, lat, lon").
		Where("id = ?", id).
		First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Или кастомная ошибка
		}
		return nil, result.Error
	}
	return &event, nil
}

func (r *eventRepository) Create(ctx context.Context, event *model.Event) (*model.Event, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Пробрасываем панику дальше
		}
	}()

	if err := tx.Omit("ParticipantsCount").Create(event).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	if len(event.Orgs) > 0 {
		if err := tx.Model(event).Association("Orgs").Append(event.Orgs); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to append orgs: %w", err)
		}
	}

	if len(event.AvailableRoles) > 0 {
		if err := tx.Model(event).Association("AvailableRoles").Append(event.AvailableRoles); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to append roles: %w", err)
		}
	}

	var created model.Event
	if err := tx.
		Preload("Orgs").
		Preload("AvailableRoles").
		Preload("Attachments").
		First(&created, "id = ?", event.ID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to load created event: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return &created, nil
}

func (r *eventRepository) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*model.Event, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingEvent model.Event
	if err := tx.First(&existingEvent, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Обновляем только переданные поля
	if len(fields) > 0 {
		if err := tx.Model(&existingEvent).Omit("ParticipantsCount").Updates(fields).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update event: %w", err)
		}
	}

	// Обновляем ассоциации только если они переданы
	if orgs != nil {
		// Используем временный объект с заполненным ID, чтобы GORM корректно понял родителя
		parent := model.Event{ID: id}
		if err := tx.Model(&parent).Association("Orgs").Replace(orgs); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to replace orgs: %w", err)
		}
	}

	if availableRoles != nil {
		// Аналогично для ролей
		parent := model.Event{ID: id}
		if err := tx.Model(&parent).Association("AvailableRoles").Replace(availableRoles); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to replace roles: %w", err)
		}
	}

	var updatedEvent model.Event
	if err := tx.
		Preload("Orgs").
		Preload("AvailableRoles").
		Preload("Attachments").
		First(&updatedEvent, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to load updated event: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return &updatedEvent, nil
}

func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Предполагаем hard-delete; если нужен soft-delete, добавь поле deleted_at
	result := tx.Delete(&model.Event{ID: id})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	return tx.Commit().Error
}

func (r *eventRepository) GetEventLocationData(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var event model.Event
	result := r.db.WithContext(ctx).
		Model(&model.Event{}).
		Select(`lat, long`).
		Where("id = ?", id).
		First(&event)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &event, nil
}

func (r *eventRepository) baseEventQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("events").
		Select(`events.id, events.name, events.status, 
			(SELECT COUNT(*) FROM event_participants ep WHERE ep.event_id = events.id) AS participants_count,
			events.starts_at`).
		Preload("Orgs", func(db *gorm.DB) *gorm.DB { return db.Select(selectUserFields) }).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.Select(`event_participants.id, event_participants.user_id, event_participants.event_id`).Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB { return db.Select(selectUserFields) })
}
