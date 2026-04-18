package repository

import (
	"context"
	"ne_noy/internal/model"
)

// EventQueueRepository
type EventQueueRepository interface {
	GetAll(ctx context.Context) ([]model.EventQueueModel, error)
	AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error
	RemovePostFromQueue(ctx context.Context, postId int64) error
}
