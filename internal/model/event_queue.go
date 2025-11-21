package model

import (
	"time"

	"github.com/google/uuid"
)

type EventQueue struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	VkPost        *string
	VkVote        *string
	VkAttachments *string
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}
