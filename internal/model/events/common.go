package events

import (
	"fmt"
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

func GetEventTableName(eventType string) (string, error) {
	switch eventType {
	case EventAsEvent:
		return "event_as_events", nil
	case EventAsTeam:
		return "event_as_teams", nil
	case EventAsTest:
		return "event_as_tests", nil
	default:
		return "", fmt.Errorf("unsupported event type: %s", eventType)
	}
}
