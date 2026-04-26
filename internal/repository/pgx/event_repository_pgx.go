package pgx

import (
	"context"
	"fmt"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/model/events/as_team"
	"ne_noy/internal/model/events/as_test"
	"ne_noy/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ne_noy/internal/model"
)

type eventRepositoryPgx struct {
	pool *pgxpool.Pool
}

func (e *eventRepositoryPgx) GetAll(ctx context.Context, roleCode *string, archived *bool) ([]*events.EventView, error) {
	condition := strings.Builder{}
	if archived == nil || *archived == false {
		condition.WriteString("ends_at < NOW()")
	} else {
		condition.WriteString("ends_at >= NOW()")
	}
	if roleCode != nil {
		condition.WriteString("AND role_code = $1")
	}
	query := fmt.Sprintf("SELECT * FROM events WHERE %s", condition.String())
	rows, err := e.pool.Query(ctx, query, roleCode)
}

func (e *eventRepositoryPgx) GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*events.EventView, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetLocationById(ctx context.Context, id uuid.UUID) (lat, long *float64, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) ExistUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetEventOrgs(ctx context.Context, id uuid.UUID, limit int) ([]model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetByVkPollId(ctx context.Context, pollId int64) (*as_event.AsEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetEventById(ctx context.Context, id uuid.UUID) (*as_event.AsEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetTeamById(ctx context.Context, id uuid.UUID) (*as_team.AsTeam, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetTestById(ctx context.Context, id uuid.UUID) (*as_test.AsTest, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) GetParticipants(ctx context.Context, id uuid.UUID, limit int) ([]as_event.EventParticipant, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) CreateEvent(ctx context.Context, event *as_event.AsEvent) (*as_event.AsEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) CreateTeam(ctx context.Context, team *as_team.AsTeam) (*as_team.AsTeam, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) CreateTest(ctx context.Context, test *as_test.AsTest) (*as_test.AsTest, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*events.EventView, error) {
	//TODO implement me
	panic("implement me")
}

func (e *eventRepositoryPgx) Delete(ctx context.Context, id uuid.UUID, eventType string) error {
	//TODO implement me
	panic("implement me")
}

func NewEventRepositoryPgx(pool *pgxpool.Pool) repository.EventRepository {
	return &eventRepositoryPgx{pool: pool}
}
