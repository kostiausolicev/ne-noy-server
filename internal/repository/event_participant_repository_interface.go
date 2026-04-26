package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// EventParticipantRepository describes persistence operations for as_event participants.
type EventParticipantRepository interface {
	// CheckParticipant marks an as_event participant as checked in.
	CheckParticipant(ctx context.Context, participant *model.EventParticipant) error

	// Participant creates a participation record for the specified VK user.
	Participant(ctx context.Context, eventID uuid.UUID, userVkId int64, prepareType string) (bool, error)

	// ParticipantById creates a participation record for the specified internal user identifier.
	ParticipantById(ctx context.Context, eventID uuid.UUID, userId uuid.UUID, prepareType string) (bool, error)

	// UnParticipant removes a participation record for the specified VK user.
	UnParticipant(ctx context.Context, eventID uuid.UUID, userId int64) (bool, error)
}
