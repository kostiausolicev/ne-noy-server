package dto

import "github.com/google/uuid"

type UserDto struct {
	ID                    uuid.UUID `json:"id"`
	VkId                  int64     `json:"vkId"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	PhotoURL              string    `json:"photo"`
	Role                  RoleDto   `json:"role"`
	GeoAvailable          bool      `json:"geoAvailable"`
	NotificationAvailable bool      `json:"notificationAvailable"`
	IsAdmin               bool      `json:"isAdmin,omitempty"`
	IsEduParticipant      bool      `json:"isEduParticipant,omitempty"`
}

type UserMiniDto struct {
	ID        uuid.UUID `json:"id"`
	VkId      int64     `json:"vk_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	PhotoURL  string    `json:"photo"`
}
