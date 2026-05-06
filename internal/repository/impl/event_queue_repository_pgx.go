package impl

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type eventQueueRepository struct {
	pool *pgxpool.Pool
}

func NewEventQueueRepository(pool *pgxpool.Pool) repository.EventQueueRepository {
	return &eventQueueRepository{pool: pool}
}

func (e *eventQueueRepository) GetAll(ctx context.Context) ([]model.EventQueueModel, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT 
		    queue_events.id, queue_events.post_id, queue_events.text, queue_events.lat, queue_events.lon, 
		    queue_events.address, queue_events.poll, queue_events.photos, queue_events.attachments, queue_events.created 
		FROM queue_events
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.EventQueueModel, 0)
	for rows.Next() {
		var m model.EventQueueModel
		if err := rows.Scan(&m.ID, &m.PostID, &m.Text, &m.Lat, &m.Lon, &m.Address, &m.Poll, &m.Photos, &m.Attachments, &m.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}

func (e *eventQueueRepository) AddPostToQueue(ctx context.Context, eventQueue *model.EventQueueModel) error {
	_, err := e.pool.Exec(ctx, `
		INSERT INTO queue_events (id, post_id, text, lat, lon, address, poll, photos, attachments)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, eventQueue.ID, eventQueue.PostID, eventQueue.Text, eventQueue.Lat, eventQueue.Lon, eventQueue.Address, eventQueue.Poll, eventQueue.Photos, eventQueue.Attachments)
	return err
}

func (e *eventQueueRepository) RemovePostFromQueue(ctx context.Context, postId int64) error {
	_, err := e.pool.Exec(ctx, `DELETE FROM queue_events WHERE post_id = $1`, postId)
	return err
}
