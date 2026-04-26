package events

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

type EventAttachment struct {
	model.BaseModel
	EventID      uuid.UUID
	EventType    string
	Event        EventProfile
	AttachmentID int64
	Attachment   model.Attachment
}
