package model

import (
	"time"

	"github.com/google/uuid"
)

type EventParticipant struct {
	ID      uuid.UUID
	EventID uuid.UUID
	Event   Event
	UserID  uuid.UUID
	User    User
	// Способ предварительной отметки на мероприятии. Через приложение или через опрос
	PrepareType    string
	IsChecked      bool
	CheckTimestamp *time.Time
	CheckLat       *float64
	CheckLong      *float64
	CheckType      string
	CheckAuthor    *uuid.UUID
	CreatedAt      time.Time
}
