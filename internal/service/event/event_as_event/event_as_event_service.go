package event_as_event

import (
	"context"
	"ne_noy/internal/dto"

	"github.com/google/uuid"
)

type EventAsEventService interface {
	GetEventById(ctx context.Context, eventId uuid.UUID) (dto.EventMiniDto, error)
	UpdateEvent(ctx context.Context, eventId uuid.UUID, event dto.CreateUpdateEventDto) (dto.EventMiniDto, error)
	CreateEvent(ctx context.Context, event dto.CreateUpdateEventDto) (dto.EventMiniDto, error)
	GetEventParticipants(ctx context.Context, eventId uuid.UUID) ([]dto.UserMiniDto, error)
}
