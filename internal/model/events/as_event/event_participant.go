package as_event

import (
	"ne_noy/internal/model"
	"time"

	"github.com/google/uuid"
)

type EventParticipants struct {
	model.BaseModel
	EventID uuid.UUID
	Event   AsEvent
	UserID  uuid.UUID
	User    model.User
	// Способ предварительной отметки на мероприятии. Через приложение или через опрос
	PrepareType    string
	IsChecked      bool
	CheckTimestamp *time.Time
	CheckLat       *float64
	CheckLong      *float64
	CheckType      string
	CheckAuthor    *uuid.UUID
}
