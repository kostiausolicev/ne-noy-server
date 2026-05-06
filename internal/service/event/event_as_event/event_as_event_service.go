package event_as_event

import (
	"context"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/event_dto"

	"github.com/google/uuid"
)

type EventAsEventService interface {
	GetEventById(ctx context.Context, eventId uuid.UUID) (dto.EventMiniDto, error)
	UpdateEvent(ctx context.Context, eventId uuid.UUID, event event_dto.CreateUpdateEventDto) (dto.EventMiniDto, error)
	CreateEvent(ctx context.Context, event event_dto.CreateUpdateEventDto) (dto.EventMiniDto, error)
	GetEventParticipants(ctx context.Context, eventId uuid.UUID) ([]dto.UserMiniDto, error)
}
