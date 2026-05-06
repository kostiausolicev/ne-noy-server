package as_team

import "ne_noy/internal/model/events"

type AsTeam struct {
	events.EventProfile
	events.EventRelations
	TeamsConstraint   int
	TeamsCapMin       *int
	TeamsCapMax       *int
	Lat               *float64
	Lon               *float64
	Address           *string
	AdditionalAddress *string
	VkPostID          *int64
}

func (e AsTeam) TableName() string {
	return "event_as_teams"
}
