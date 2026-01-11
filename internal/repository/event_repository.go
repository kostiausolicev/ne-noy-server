package repository

import (
	"fmt"
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	selectUserFields = `
				"users".id, 
				"users".vk_id,
				"users".first_name, 
				"users".last_name,
				"users".photo_url
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
	GetEventLocationData(id uuid.UUID) (*model.Event, error)
	GetUserParticipationInEvent(eventId uuid.UUID, userId int64) (bool, error)
	GetEventOrgs(eventId uuid.UUID) ([]model.User, error)
	GetByVkPollAnswerId(answerId int64) (*model.Event, error)
	GetById(id uuid.UUID) (*model.Event, error)
	GetParticipants(id uuid.UUID) ([]model.EventParticipant, error)
	Create(event *model.Event) (*model.Event, error)
	Update(event *model.Event) (*model.Event, error)
	Delete(id uuid.UUID) error
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
		Joins("LEFT JOIN event_orgs eo ON eo.user_id = ?", orgId).
		Where("eo.user_id = ?", orgId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetUserParticipationInEvent(eventId uuid.UUID, userVkId int64) (bool, error) {
	var exists bool

	err := r.db.
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

func (r *eventRepository) GetByVkPollAnswerId(vkPollAnswerId int64) (*model.Event, error) {
	var event model.Event
	result := r.db.Table("events e").
		Select("e.id").
		Where("e.vk_poll_answer_id = ?", vkPollAnswerId).
		Scan(&event)
	return &event, result.Error
}

func (r *eventRepository) GetEventOrgs(eventId uuid.UUID) ([]model.User, error) {
	orgs := make([]model.User, 0)

	err := r.db.
		Table("event_orgs eo").
		Select("u.id, u.vk_id").
		Joins("JOIN users u ON u.id = eo.user_id").
		Where("eo.event_id = ?", eventId).
		Scan(&orgs).Error

	if err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *eventRepository) GetAllByRole(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := r.getEventsQuery().
		Joins("JOIN event_roles er ON er.event_id = events.id").
		Where("er.role_id = ? AND events.starts_at > NOW()", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetAllArchive(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := r.getEventsQuery().
		Joins("JOIN event_roles er ON er.event_id = events.id").
		Where("er.role_id = ? AND events.starts_at < NOW()", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *eventRepository) GetParticipants(id uuid.UUID) ([]model.EventParticipant, error) {
	var participants []model.EventParticipant

	result := r.db.
		Table("event_participants").
		Select(`
			event_participants.user_id,
			event_participants.check_timestamp,
			event_participants.is_checked
		`).
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

func (r *eventRepository) GetById(id uuid.UUID) (*model.Event, error) {
	var event *model.Event
	result := r.db.
		Table("events").
		Preload("Orgs", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.
				Select(`
					event_participants.id, 
					event_participants.user_id,
					event_participants.event_id
				`).
				Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB {
			return db.Select(selectUserFields)
		}).
		Preload("Attachments", func(db *gorm.DB) *gorm.DB {
			return db.Select(`
				event_attachments.id,
				event_attachments.attachment_link
			`)
		}).
		Select(`
			events.id,
			events.name,
			events.cover,
			events.description,
			events.address,
			events.vk_post_id,
			events.vk_vote_id,
			events.status,
			events.starts_at
		`).
		Where("events.id = ?", id).
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
	if event.VkPostId != nil {
		existingEvent.VkPostId = event.VkPostId
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

	if err := tx.Omit("ParticipantsCount").Save(&existingEvent).Error; err != nil {
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
		Table("events").
		Select(`
			events.lat,
			events.long
		`).
		Where("events.id = ?", id).
		Take(event)
	if result.Error != nil {
		return nil, result.Error
	}
	return event, nil
}

func (r *eventRepository) getEventsQuery() *gorm.DB {
	return r.db.
		Table("events").
		Select(`
            events.id,
            events.name,
			events.status,
			(
				SELECT COUNT(*)
				FROM event_participants ep
				WHERE ep.event_id = events.id
			) AS participants_count,
            events.starts_at
        `).
		Preload("Orgs", func(db *gorm.DB) *gorm.DB { return db.Select(selectUserFields) }).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.
				Select(`
					event_participants.id, 
					event_participants.user_id,
					event_participants.event_id
				`).
				Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB { return db.Select(selectUserFields) })
}
