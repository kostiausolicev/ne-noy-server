package repository

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	select_user_fields = `
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

	result := e.getEventsQuery().
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (e eventRepository) GetAllByOrg(orgId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := e.getEventsQuery().
		Joins("LEFT JOIN event_org eo ON eo.user_id = ?", orgId).
		Where("eo.user_id = ?", orgId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (e eventRepository) GetUserParticipationInEvent(eventId uuid.UUID, userVkId int64) (bool, error) {
	var count int64
	result := e.db.
		Table("event_participant").
		Joins(`INNER JOIN "user" on event_participant.user_id = "user".id`).
		Where(`event_id = ? AND "user".vk_id = ?`, eventId, userVkId).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (e eventRepository) GetAllByRole(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := e.getEventsQuery().
		Joins("JOIN event_role er ON er.event_id = event.id").
		Where("er.role_id = ? AND event.starts_at > NOW()", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (e eventRepository) GetAllArchive(roleId uuid.UUID) ([]*model.Event, error) {
	var events []*model.Event

	result := e.getEventsQuery().
		Joins("JOIN event_role er ON er.event_id = event.id").
		Where("er.role_id = ? AND event.starts_at < NOW()", roleId).
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (e eventRepository) GetParticipants(id uuid.UUID) ([]model.EventParticipant, error) {
	var participants []model.EventParticipant

	result := e.db.
		Table("event_participant").
		Select(`
			event_participant.user_id,
			event_participant.check_timestamp,
			event_participant.is_checked
		`).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(select_user_fields)
		}).
		Where("event_participant.event_id = ?", id).
		Find(&participants)

	if result.Error != nil {
		return nil, result.Error
	}
	return participants, nil
}

// GetById TODO Попробовать использовать Joins вместе со Scan
func (e eventRepository) GetById(id uuid.UUID) (*model.Event, error) {
	var event *model.Event
	result := e.db.
		Table("event").
		Preload("Orgs", func(db *gorm.DB) *gorm.DB {
			return db.Select(`
				"user".id, 
				"user".vk_id,
				"user".first_name, 
				"user".last_name,
				"user".photo_url
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
				"user".vk_id,
				"user".first_name,
				"user".last_name,
				"user".photo_url
			`)
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

func (e eventRepository) GetEventLocationData(id uuid.UUID) (*model.Event, error) {
	var event *model.Event
	result := e.db.
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

func (e eventRepository) getEventsQuery() *gorm.DB {
	return e.db.
		Table("event").
		Select(`
            event.id,
            event.name,
            event.starts_at
        `).
		Preload("Orgs", func(db *gorm.DB) *gorm.DB { return db.Select(select_user_fields) }).
		Preload("EventParticipants", func(db *gorm.DB) *gorm.DB {
			return db.
				Select(`
					event_participant.id, 
					event_participant.user_id,
					event_participant.event_id
				`).
				Limit(3)
		}).
		Preload("EventParticipants.User", func(db *gorm.DB) *gorm.DB { return db.Select(select_user_fields) })
}
