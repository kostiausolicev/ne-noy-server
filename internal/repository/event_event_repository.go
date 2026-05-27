package repository

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"

	"github.com/google/uuid"
)

type EventEventRepository interface {
	GetLocationById(ctx context.Context, id uuid.UUID) (lat, long *float64, err error)

	ExistUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) (bool, error)

	GetEventOrgs(ctx context.Context, id uuid.UUID, limit int) ([]model.User, error)

	GetByVkPollId(ctx context.Context, pollId int64) (*as_event.AsEvent, error)

	GetEventById(ctx context.Context, id uuid.UUID) (*as_event.AsEvent, error)

	GetParticipants(ctx context.Context, id uuid.UUID, limit int) ([]as_event.EventParticipants, error)

	CreateEvent(ctx context.Context, event *as_event.AsEvent) (*as_event.AsEvent, error)

	Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, roleCodes []string, attachments []events.EventAttachment) (*as_event.AsEvent, error)
}
