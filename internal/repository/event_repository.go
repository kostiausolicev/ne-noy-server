package repository

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

type EventRepository interface {
	GetAll() ([]*model.Event, error)
	GetAllByRole(roleId uuid.UUID) ([]*model.Event, error)
	CountParticipants(id uuid.UUID) (int, error)

	GetById(id uuid.UUID) (*model.Event, error)
	Create(event *model.Event) (*model.Event, error)
	Update(event *model.Event) (*model.Event, error)
	Delete(id uuid.UUID) error
}

func (e eventRepository) CountParticipants(id uuid.UUID) (int, error) {
	var count int64
	result := e.db.
		Table("event_participant").
		Where("event_id = ?", id).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return int(count), nil
}

func (e eventRepository) GetAll() ([]*model.Event, error) {
	var events []*model.Event

	result := e.db.
		Table("event").
		Select(`
            event.id,
            event.name,
            event.starts_at
        `).
		Preload("Orgs", func(db *gorm.DB) *gorm.DB {
			return db.Select(`
				"user".id, 
				"user".first_name, 
				"user".last_name
			`)
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
			return db.Select(`
				"user".id,
				"user".first_name,
				"user".last_name
			`)
		}).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (e eventRepository) GetAllByRole(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := e.db.
		Table("event").
		Preload("Attachments").
		Preload("Orgs", func(db *gorm.DB) *gorm.DB {
			return db.Select(`
				"user".id, 
				"user".vk_id, 
				"user".first_name, 
				"user".last_name
			`)
		}).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.Select("event_participant.id", "event_participant.user_id", "event_participant.event_id")
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB {
			return db.Select(`
				"user".id, 
				"user".vk_id, 
				"user".first_name, 
				"user".last_name
			`)
		}).
		Joins("JOIN event_role er ON er.event_id = events.id").
		Where("er.role_id = ?", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (e eventRepository) GetById(id uuid.UUID) (*model.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventRepository) Create(event *model.Event) (*model.Event, error) {
	result := e.db.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Create(&event)
	if result.Error != nil {
		return nil, result.Error
	}
	return event, nil
}

func (e eventRepository) Update(event *model.Event) (*model.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (e eventRepository) Delete(id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}
