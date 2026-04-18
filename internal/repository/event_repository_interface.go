package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// EventRepository defines event related operations
type EventRepository interface {
	GetAll(ctx context.Context) ([]*model.Event, error)
	GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*model.Event, error)
	GetAllByRole(ctx context.Context, role string) ([]*model.Event, error)
	GetAllArchive(ctx context.Context, role string) ([]*model.Event, error)
	GetEventLocationData(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userVkId int64) (bool, error)
	GetEventOrgs(ctx context.Context, eventId uuid.UUID) ([]model.User, error)
	GetByVkPollId(ctx context.Context, pollId int64) (*model.Event, error)
	GetById(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetLocationById(ctx context.Context, id uuid.UUID) (*model.Event, error)
	GetParticipants(ctx context.Context, id uuid.UUID) ([]model.EventParticipant, error)
	Create(ctx context.Context, event *model.Event) (*model.Event, error)
	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*model.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
