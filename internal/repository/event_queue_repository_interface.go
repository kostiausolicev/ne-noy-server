package repository

import (
	"context"
	"ne_noy/internal/model"
)

// EventQueueRepository describes persistence operations for the as_event post queue.
type EventQueueRepository interface {
	// GetAll returns all queued posts.
	GetAll(ctx context.Context) ([]model.EventQueueModel, error)

	// AddPostToQueue adds a post to the publishing queue.
	AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error

	// RemovePostFromQueue removes a post from the queue by VK post identifier.
	RemovePostFromQueue(ctx context.Context, postId int64) error
}
