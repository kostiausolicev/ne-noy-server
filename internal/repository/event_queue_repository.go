package repository

import (
	"context"
	"ne_noy/internal/model"

	"gorm.io/gorm"
)

type EventQueueRepository interface {
	GetAll(ctx context.Context) ([]model.EventQueueModel, error)
	AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error
	RemovePostFromQueue(ctx context.Context, postId int64) error
}

func (e *eventQueueRepository) withScope(ctx context.Context) *gorm.DB {
	return e.db.WithContext(ctx)
}

type eventQueueRepository struct {
	db *gorm.DB
}

func (e *eventQueueRepository) GetAll(ctx context.Context) ([]model.EventQueueModel, error) {
	var result []model.EventQueueModel
	res := e.withScope(ctx).
		Table("queue_events").
		Select("queue_events.*").
		Find(&result)

	return result, res.Error
}

func (e *eventQueueRepository) AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error {
	result := e.withScope(ctx).
		Create(eventQueue)
	return result.Error
}

func (e *eventQueueRepository) RemovePostFromQueue(ctx context.Context, postId int64) error {
	return e.withScope(ctx).Delete(&model.EventQueueModel{PostID: postId}).Error
}

func NewEventQueueRepository(db *gorm.DB) EventQueueRepository {
	return &eventQueueRepository{db: db}
}
