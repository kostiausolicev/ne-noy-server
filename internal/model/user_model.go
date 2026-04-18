package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID
	VkID                  int64
	FirstName             string
	LastName              string
	RoleID                *uuid.UUID
	Role                  *Role
	PhotoURL              string
	GeoAvailable          bool
	NotificationAvailable bool
	CreatedAt             time.Time
}

func (e User) TableName() string {
	return "users"
}
