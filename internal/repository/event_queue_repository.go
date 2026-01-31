package repository

import (
	"context"
	"ne_noy/internal/model"

	"gorm.io/gorm"
)

type EventQueueRepository interface {
	AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error
	RemovePostFromQueue(ctx context.Context, postId int64) error
}

func (e *eventQueueRepository) withScope(ctx context.Context) *gorm.DB {
	return e.db.WithContext(ctx)
}

type eventQueueRepository struct {
	db *gorm.DB
}

func (e *eventQueueRepository) AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error {
	result := e.withScope(ctx).
		Create(eventQueue)
	return result.Error
}

func (e *eventQueueRepository) RemovePostFromQueue(ctx context.Context, postId int64) error {
	//TODO implement me
	panic("implement me")
}

func NewEventQueueRepository(db *gorm.DB) EventQueueRepository {
	return &eventQueueRepository{db: db}
}
