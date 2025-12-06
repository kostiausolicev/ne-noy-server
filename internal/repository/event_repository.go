package repository

import (
	"fmt"
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	selectUserFields = `
				"user".id, 
				"user".vk_id,
				"user".first_name, 
				"user".last_name,
				"user".photo_url
			`
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

type EventRepository interface {
	GetAll() ([]*model.Event, error)
	GetAllByOrg(orgId uuid.UUID) ([]*model.Event, error)
	GetAllByRole(roleId uuid.UUID) ([]*model.Event, error)
	GetAllArchive(roleId uuid.UUID) ([]*model.Event, error)
	CountParticipants(id uuid.UUID) (int, error)
	GetEventLocationData(id uuid.UUID) (*model.Event, error)
	GetUserParticipationInEvent(eventId uuid.UUID, userId int64) (bool, error)

	GetById(id uuid.UUID) (*model.Event, error)
	GetParticipants(id uuid.UUID) ([]model.EventParticipant, error)
	Create(event *model.Event) (*model.Event, error)
	Update(event *model.Event) (*model.Event, error)
	Delete(id uuid.UUID) error
}

func (r *eventRepository) CountParticipants(id uuid.UUID) (int, error) {
	var count int64
	result := r.db.
		Table("event_participant").
		Where("event_id = ?", id).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(count), nil
}

func (r *eventRepository) GetAll() ([]*model.Event, error) {
	var events []*model.Event

	result := r.getEventsQuery().
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetAllByOrg(orgId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := r.getEventsQuery().
		Joins("LEFT JOIN event_org eo ON eo.user_id = ?", orgId).
		Where("eo.user_id = ?", orgId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetUserParticipationInEvent(eventId uuid.UUID, userVkId int64) (bool, error) {
	var count int64
	result := r.db.
		Table("event_participant").
		Joins(`INNER JOIN "user" on event_participant.user_id = "user".id`).
		Where(`event_id = ? AND "user".vk_id = ?`, eventId, userVkId).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (r *eventRepository) GetAllByRole(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := r.getEventsQuery().
		Joins("JOIN event_role er ON er.event_id = event.id").
		Where("er.role_id = ? AND event.starts_at > NOW()", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetAllArchive(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := r.getEventsQuery().
		Joins("JOIN event_role er ON er.event_id = event.id").
		Where("er.role_id = ? AND event.starts_at < NOW()", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetParticipants(id uuid.UUID) ([]model.EventParticipant, error) {
	var participants []model.EventParticipant

	result := r.db.
		Table("event_participant").
		Select(`
			event_participant.user_id,
			event_participant.check_timestamp,
			event_participant.is_checked
		`).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Where("event_participant.event_id = ?", id).
		Find(&participants)

	if result.Error != nil {
		return nil, result.Error
	}
	return participants, nil
}

func (r *eventRepository) GetById(id uuid.UUID) (*model.Event, error) {
	var event *model.Event
	result := r.db.
		Table("event").
		Preload("Orgs", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.
				Select(`
					event_participant.id, 
					event_participant.user_id,
					event_participant.event_id
				`).
				Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Select(`
				event_attachment.id,
				event_attachment.attachment_link
			`)
		}).
		Select(`
			event.id,
			event.name,
			event.cover,
			event.description,
			event.address,
			event.vk_post_id,
			event.vk_vote_id,
			event.address,
			event.starts_at
		`).
		Where("event.id = ?", id).
		Find(&event)

	if result.Error != nil {
		return nil, result.Error
	}
	return event, nil
}

func (r *eventRepository) Create(event *model.Event) (*model.Event, error) {
	tx := r.db.Begin()

	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	// 1. сохраняем сам event
	if err := tx.Create(event).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	// 2. добавляем Orgs
	if len(event.Orgs) > 0 {
		if err := tx.Model(event).Association("Orgs").Append(event.Orgs); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to append orgs: %w", err)
		}
	}

	// 3. добавляем роли
	if len(event.AvailableRoles) > 0 {
		if err := tx.Model(event).Association("AvailableRoles").Append(event.AvailableRoles); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to append roles: %w", err)
		}
	}

	// 4. подгружаем связанные данные
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

func (r *eventRepository) Update(event *model.Event) (*model.Event, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingEvent model.Event
	if err := tx.First(&existingEvent, "id = ?", event.ID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.Name != "" {
		existingEvent.Name = event.Name
	}
	if event.Description != nil {
		existingEvent.Description = event.Description
	}
	if event.Cover != nil {
		existingEvent.Cover = event.Cover
	}
	if event.VkPostLink != nil {
		existingEvent.VkPostLink = event.VkPostLink
	}
	if event.Address != nil {
		existingEvent.Address = event.Address
	}
	if event.Lat != nil {
		existingEvent.Lat = event.Lat
	}
	if event.Long != nil {
		existingEvent.Long = event.Long
	}
	if event.Status != nil {
		existingEvent.Status = event.Status
	}
	if event.StartsAt != nil {
		existingEvent.StartsAt = event.StartsAt
	}

	if err := tx.Model(&existingEvent).Association("Orgs").Clear(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to clear orgs: %w", err)
	}

	if err := tx.Model(&existingEvent).Association("AvailableRoles").Clear(); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to clear roles: %w", err)
	}

	if err := tx.Save(&existingEvent).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	if len(event.Orgs) > 0 {
		if err := tx.Model(&existingEvent).Association("Orgs").Append(&event.Orgs); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to append orgs: %w", err)
		}
	}

	if len(event.AvailableRoles) > 0 {
		if err := tx.Model(&existingEvent).Association("AvailableRoles").Append(&event.AvailableRoles); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to append roles: %w", err)
		}
	}

	var updatedEvent model.Event
	if err := tx.Preload("Orgs").
		Preload("AvailableRoles").
		Preload("Attachments").
		First(&updatedEvent, "id = ?", event.ID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to load updated event: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return &updatedEvent, nil
}

func (r *eventRepository) Delete(id uuid.UUID) error {
	panic("implement me")
}

func (r *eventRepository) GetEventLocationData(id uuid.UUID) (*model.Event, error) {
	var event *model.Event
	result := r.db.
		Table("event").
		Select(`
			event.lat,
			event.long
		`).
		Where("event.id = ?", id).
		Take(event)
	if result.Error != nil {
		return nil, result.Error
	}
	return event, nil
}

func (r *eventRepository) getEventsQuery() *gorm.DB {
	return r.db.
		Table("event").
		Select(`
            event.id,
            event.name,
            event.starts_at
        `).
		Preload("Orgs", func(db *gorm.DB) *gorm.DB { return db.Select(selectUserFields) }).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.
				Select(`
					event_participant.id, 
					event_participant.user_id,
					event_participant.event_id
				`).
				Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB { return db.Select(selectUserFields) })
}
