package service

import (
	"context"
	"errors"
	"ne_noy/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventQueueServiceDeletePostFromQueueCallsRepo(t *testing.T) {
	ctx := context.Background()
	repo := newFakeQueueRepo()
	repo.posts[42] = model.EventQueueModel{PostID: 42}
	svc := NewEventQueueService(repo)

	require.NoError(t, svc.DeletePostFromQueue(ctx, 42))
	_, exists := repo.posts[42]
	require.False(t, exists)
}

func TestEventQueueServiceDeletePostFromQueueReturnsErrorOnMissing(t *testing.T) {
	ctx := context.Background()
	repo := newFakeQueueRepo()
	svc := NewEventQueueService(repo)

	err := svc.DeletePostFromQueue(ctx, 999)
	require.Error(t, err)
}

type fakeQueueRepo struct {
	posts map[int64]model.EventQueueModel
}

func newFakeQueueRepo() *fakeQueueRepo {
	return &fakeQueueRepo{posts: make(map[int64]model.EventQueueModel)}
}

func (f *fakeQueueRepo) GetAll(_ context.Context) ([]model.EventQueueModel, error) {
	result := make([]model.EventQueueModel, 0, len(f.posts))
	for _, p := range f.posts {
		result = append(result, p)
	}
	return result, nil
}

func (f *fakeQueueRepo) AddPostToQueue(_ context.Context, m *model.EventQueueModel) error {
	f.posts[m.PostID] = *m
	return nil
}

func (f *fakeQueueRepo) RemovePostFromQueue(_ context.Context, postID int64) error {
	if _, ok := f.posts[postID]; !ok {
		return errors.New("post not found")
	}
	delete(f.posts, postID)
	return nil
}
