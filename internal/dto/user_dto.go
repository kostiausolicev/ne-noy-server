package dto

import "github.com/google/uuid"

type UserDto struct {
	ID                    uuid.UUID `json:"id"`
	VkId                  int64     `json:"vk_id"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	PhotoURL              string    `json:"photo"`
	Role                  RoleDto   `json:"role"`
	GeoAvailable          bool      `json:"geo_available"`
	NotificationAvailable bool      `json:"notification_available"`
}

type UserMiniDto struct {
	ID        uuid.UUID `json:"id"`
	VkId      int64     `json:"vk_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	PhotoURL  string    `json:"photo"`
}
