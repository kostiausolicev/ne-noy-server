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
	Attachments    []EventAttachment
	Orgs           []model.User
}
