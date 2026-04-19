package model

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID                uuid.UUID
	Name              string
	EventType         string
	Status            *string
	Description       *string
	Cover             *string
	VkPostId          *int64
	VkVoteID          *int64
	VkPollAnswerID    *int64
	Lat               *float64
	Long              *float64
	Address           *string
	AdditionalAddress *string
	StartsAt          *time.Time
	EndsAt            *time.Time
	ParticipantsCount int
	Type              *string
	CreatedAt         time.Time

	AvailableRoles    []Role
	Attachments       []EventAttachment
	Orgs              []User
	EventParticipants []EventParticipant
}

func (e Event) TableName() string {
	return "events"
}

const (
	EventTypeEvent    = "event"
	EventTypeActivity = "activity"
	EventTypeTeam     = "team"
	EventTypePoll     = "poll"
	EventTypeTest     = "test"
)
