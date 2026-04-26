package model

import (
	"github.com/google/uuid"
)

type User struct {
	BaseModel
	VkID                  int64
	FirstName             string
	LastName              string
	RoleID                *uuid.UUID
	Role                  *Role
	PhotoURL              string
	GeoAvailable          bool
	NotificationAvailable bool
}

func (e User) TableName() string {
	return "users"
}
