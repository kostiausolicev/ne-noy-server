package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	Title     string
	Fragment  *string
	SendAt    time.Time  `gorm:"column:sendAt"`
	UserRole  *uuid.UUID `gorm:"column:userRole"`
	Role      *Role
	ForAll    bool      `gorm:"column:forAll"`
	CreatedAt time.Time `gorm:"column:createdAt"`

	Users []NotificationUser
}

func (n Notification) TableName() string {
	return "notifications"
}

type NotificationUser struct {
	ID             uuid.UUID
	NotificationID *uuid.UUID `gorm:"column:notificationid"`
	Notification   *Notification
	UserID         *uuid.UUID `gorm:"column:userid"`
	User           *User
	Read           bool
}

func (n NotificationUser) TableName() string {
	return "notification_user"
}
