package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserByLinksDto struct {
	Links []string `json:"links"`
}

type CreateUserDto struct {
	VkId      int64  `json:"vkId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	PhotoURL  string `json:"photo"`
}

type UserDto struct {
	ID                    *uuid.UUID `json:"id,omitempty"`
	VkId                  int64      `json:"vkId"`
	FirstName             string     `json:"firstName"`
	LastName              string     `json:"lastName"`
	PhotoURL              string     `json:"photo"`
	Role                  RoleDto    `json:"role"`
	GeoAvailable          bool       `json:"geoAvailable"`
	NotificationAvailable bool       `json:"notificationAvailable"`
	IsAdmin               bool       `json:"isAdmin,omitempty"`
	IsEduParticipant      bool       `json:"isEduParticipant,omitempty"`
}

type UserMiniDto struct {
	ID        uuid.UUID `json:"id"`
	VkId      int64     `json:"vk_id"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	PhotoURL  string    `json:"photo"`
}

type EventParticipantDto struct {
	User           UserMiniDto  `json:"user"`
	IsChecked      bool         `json:"isChecked"`
	CheckAuthor    *UserMiniDto `json:"checkAuthor,omitempty"`
	CheckTimestamp *time.Time   `json:"checkTimestamp,omitempty"`
}

type CheckEventParticipant struct {
	UserId          uuid.UUID `json:"userId"`
	EventId         uuid.UUID `json:"eventId"`
	CheckAuthorVkId *int64    `json:"checkAuthorVkId"`
	Timestamp       time.Time `json:"timestamp"`
	CheckType       string    `json:"checkType"`
	Lat             *float64  `json:"lat"`
	Long            *float64  `json:"long"`
}
