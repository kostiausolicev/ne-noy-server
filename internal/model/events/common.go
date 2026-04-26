package events

import (
	"ne_noy/internal/model"
	"time"
)

const (
	EventAsEvent = "as_event"
	EventAsTest  = "as_test"
	EventAsTeam  = "as_team"
)

type EventProfile struct {
	model.BaseModel
	Name        string
	Description *string
	Cover       *string
	Status      string
	StartsAt    time.Time
	EndsAt      *time.Time
}

type EventRelations struct {
	Orgs        []model.User
	Attachments []EventAttachment
}
