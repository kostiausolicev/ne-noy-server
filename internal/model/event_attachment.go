package model

import (
	"time"

	"github.com/google/uuid"
)

type EventAttachment struct {
	ID           uuid.UUID
	EventID      uuid.UUID
	EventType    string
	Event        Event
	AttachmentID int64
	Attachment   Attachment
	CreatedAt    time.Time
}
