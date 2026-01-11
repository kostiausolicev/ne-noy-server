package repository

import (
	"ne_noy/internal/model"

	"gorm.io/gorm"
)

type EventQueueRepository interface {
	AddPostToQueue(eventQueue *model.EventQueueModel) error
	RemovePostFromQueue(postId int64) error
}

type eventQueueRepository struct {
	db *gorm.DB
}

func (e eventQueueRepository) AddPostToQueue(eventQueue *model.EventQueueModel) error {
	result := e.db.Create(eventQueue)
	return result.Error
}

func (e eventQueueRepository) RemovePostFromQueue(postId int64) error {
	//TODO implement me
	panic("implement me")
}

func NewEventQueueRepository(db *gorm.DB) EventQueueRepository {
	return &eventQueueRepository{db: db}
}
