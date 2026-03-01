package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	VkID                  int64
	FirstName             string     `gorm:"size:100;not null"`
	LastName              string     `gorm:"size:100;not null"`
	RoleID                *uuid.UUID `gorm:"type:uuid"`
	Role                  *Role      `gorm:"foreignKey:RoleID"`
	PhotoURL              string
	GeoAvailable          bool      `gorm:"default:false"`
	NotificationAvailable bool      `gorm:"default:true"`
	CreatedAt             time.Time `gorm:"autoCreateTime"`
}

func (e User) TableName() string {
	return "users"
}
