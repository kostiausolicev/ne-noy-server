package model

import (
	"time"

	"github.com/google/uuid"
)

type EventAttachment struct {
	ID           uuid.UUID
	EventID      uuid.UUID
	Event        Event
	AttachmentID uuid.UUID
	Attachment   Attachment
	CreatedAt    time.Time
}
