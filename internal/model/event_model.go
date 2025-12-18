package model

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name              string    `gorm:"size:255;not null"`
	Status            *string   `gorm:"size:50"` // deleted | draft | active
	Description       *string
	Cover             *string
	VkPostLink        *string `gorm:"column:vk_post_id"`
	VkVoteID          *string
	Lat               *float64 `gorm:"type:decimal(10,8)"`
	Long              *float64 `gorm:"column:lon;type:decimal(11,8)"`
	Address           *string
	StartsAt          *time.Time
	ParticipantsCount int       `gorm:"-:create"`
	Type              *string   `gorm:"size:100"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`

	AvailableRoles    []Role            `gorm:"many2many:event_roles"`
	Attachments       []EventAttachment `gorm:"many2many:event_attachments"`
	Orgs              []User            `gorm:"many2many:event_orgs"`
	EventParticipants []EventParticipant
}
