package events

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

type EventOrg struct {
	EventID   uuid.UUID
	EventType string
	UserID    uuid.UUID
	User      model.User
}

func (e EventOrg) TableName() string {
	return "event_orgs"
}

type EventRole struct {
	EventID   uuid.UUID
	EventType string
	RoleID    uuid.UUID
	Role      model.Role
}

func (e EventRole) TableName() string {
	return "event_roles"
}
