package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// EventParticipantRepository defines methods
type EventParticipantRepository interface {
	CheckParticipant(ctx context.Context, participant *model.EventParticipant) error
	Participant(ctx context.Context, eventID uuid.UUID, userVkId int64, prepareType string) (bool, error)
	ParticipantById(ctx context.Context, eventID uuid.UUID, userId uuid.UUID, prepareType string) (bool, error)
	UnParticipant(ctx context.Context, eventID uuid.UUID, userId int64) (bool, error)
}
