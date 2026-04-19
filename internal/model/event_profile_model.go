package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventProfile struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Cover       *string
	Status      string
	StartsAt    time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventRelations struct {
	Orgs              []User
	Attachments       []EventAttachment
	EventParticipants []EventParticipant
	ParticipantsCount int
}

type EventAsEvent struct {
	EventProfile
	EventRelations
	VkPostID          *int64
	VkVoteID          *int64
	VkPollAnswerID    *int64
	Lat               *float64
	Lon               *float64
	Address           *string
	AdditionalAddress *string
}

func (EventAsEvent) TableName() string {
	return "event_as_events"
}

type EventAsTest struct {
	EventProfile
	EventRelations
	ExtLinkID *string
	Attempts  int
	EventID   *uuid.UUID
	VkPostID  *int64
}

func (EventAsTest) TableName() string {
	return "event_as_tests"
}

type EventAsPoll struct {
	EventProfile
	EventRelations
	ExtLinkID *string
	VkPostID  *int64
}

func (EventAsPoll) TableName() string {
	return "event_as_polls"
}

type EventAsTeam struct {
	EventProfile
	EventRelations
	TeamsConstraint   int
	TeamsCapMin       *int
	TeamsCapMax       *int
	Lat               *float64
	Lon               *float64
	Address           *string
	AdditionalAddress *string
	VkPostID          *int64
}

func (EventAsTeam) TableName() string {
	return "event_as_teams"
}

type EventAsActivity struct {
	EventProfile
	EventRelations
	TrainParams         json.RawMessage // []string
	AvailableActivities json.RawMessage // []string
}

func (EventAsActivity) TableName() string {
	return "event_as_activities"
}
