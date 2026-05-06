package events

import (
	"ne_noy/internal/model"
	"time"
)

type EventView struct {
	model.BaseModel
	Name     string
	Status   string
	StartsAt time.Time
	EndsAt   *time.Time
	Type     string

	AvailableRoles []model.Role
	Orgs           []model.User
	Participants   []model.User
}

func (e EventView) TableName() string {
	return "events"
}
