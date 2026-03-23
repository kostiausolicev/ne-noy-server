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
	VkPostId          *int64   `gorm:"column:vk_post_id"`
	VkVoteID          *int64   `gorm:"column:vk_vote_id"`
	VkPollAnswerID    *int64   `gorm:"column:vk_poll_answer_id"`
	Lat               *float64 `gorm:"type:decimal(10,8)"`
	Long              *float64 `gorm:"column:lon;type:decimal(11,8)"`
	Address           *string
	AdditionalAddress *string
	StartsAt          *time.Time
	EndsAt            *time.Time
	ParticipantsCount int       `gorm:"->"`
	Type              *string   `gorm:"size:100"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`

	AvailableRoles    []Role            `gorm:"many2many:event_roles"`
	Attachments       []EventAttachment `gorm:"many2many:event_attachments"`
	Orgs              []User            `gorm:"many2many:event_orgs"`
	EventParticipants []EventParticipant
}

func (e Event) TableName() string {
	return "events"
}
