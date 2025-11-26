package model

import (
	"time"

	"github.com/google/uuid"
)

type EventAttachment struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	EventID        uuid.UUID `gorm:"type:uuid"`
	Event          Event     `gorm:"foreignKey:EventID"`
	AttachmentLink string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}
