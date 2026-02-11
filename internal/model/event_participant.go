package model

import (
	"time"

	"github.com/google/uuid"
)

type EventParticipant struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	EventID uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_event_user"`
	Event   Event     `gorm:"foreignKey:EventID"`
	UserID  uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_event_user"`
	User    User      `gorm:"foreignKey:UserID"`
	// Способ предварительной отметки на мероприятии. Через приложение или через опрос
	PrepareType    string `gorm:"size:5"` // vote | app
	IsChecked      bool   `gorm:"default:false"`
	CheckTimestamp *time.Time
	CheckLat       *float64   `gorm:"type:decimal(10,8)"`
	CheckLong      *float64   `gorm:"column:check_lon;type:decimal(11,8)"`
	CheckType      string     `gorm:"size:50"`
	CheckAuthor    *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
}
