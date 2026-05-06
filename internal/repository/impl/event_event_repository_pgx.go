package impl

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventEventRepositoryPgx struct {
	pool *pgxpool.Pool
}

func (e *eventEventRepositoryPgx) GetLocationById(ctx context.Context, id uuid.UUID) (lat, long *float64, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) ExistUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) GetEventOrgs(ctx context.Context, id uuid.UUID, limit int) ([]model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) GetByVkPollId(ctx context.Context, pollId int64) (*as_event.AsEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) GetEventById(ctx context.Context, id uuid.UUID) (*as_event.AsEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) GetParticipants(ctx context.Context, id uuid.UUID, limit int) ([]as_event.EventParticipants, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) CreateEvent(ctx context.Context, event *as_event.AsEvent) (*as_event.AsEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventEventRepositoryPgx) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*events.EventView, error) {
	//TODO implement me
	panic("implement me")
}

func NewEventEventRepository(db *pgxpool.Pool) repository.EventEventRepository {
	return &eventEventRepositoryPgx{pool: db}
}
