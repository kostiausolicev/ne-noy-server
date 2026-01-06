package dto

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model UserDto
type UserDto struct {
	// ID Пользователя
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id"`
	// VK ID Пользователя
	// required: true
	VkId int64 `json:"vkId"`
	// Имя Пользователя
	// required: true
	FirstName string `json:"firstName"`
	// Фамилия Пользователя
	// required: true
	LastName string `json:"lastName"`
	// Фото Пользователя
	// required: true
	PhotoURL string `json:"photo"`
	// Роль Пользователя
	// required: true
	Role RoleDto `json:"role"`
	// Доступна ли геолокация
	// required: true
	GeoAvailable bool `json:"geoAvailable"`
	// Доступны ли уведомления
	// required: true
	NotificationAvailable bool `json:"notificationAvailable"`
	// Является ли администратором
	// required: false
	IsAdmin bool `json:"isAdmin,omitempty"`
	// Является ли участником образовательной программы
	// required: false
	IsEduParticipant bool `json:"isEduParticipant,omitempty"`
}

// swagger:model UserMiniDto
type UserMiniDto struct {
	// ID Пользователя
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id"`
	// VK ID Пользователя
	// required: true
	VkId int64 `json:"vk_id"`
	// Имя Пользователя
	// required: true
	FirstName string `json:"firstname"`
	// Фамилия Пользователя
	// required: true
	LastName string `json:"lastname"`
	// Фото Пользователя
	// required: true
	PhotoURL string `json:"photo"`
}

// swagger:model EventParticipantDto
type EventParticipantDto struct {
	// Пользователь
	// required: true
	User UserMiniDto `json:"user"`
	// Присутствие на событии подтверждено
	// required: true
	IsChecked bool `json:"isChecked"`
	// Автор отметки
	// required: false
	CheckAuthor    *UserMiniDto `json:"checkAuthor,omitempty"`
	CheckTimestamp *time.Time   `json:"checkTimestamp,omitempty"`
}

// swagger:model CheckEventParticipant
type CheckEventParticipant struct {
	// ID Пользователя
	// format: uuid
	// required: true
	UserId uuid.UUID `json:"userId"`
	// ID События
	// format: uuid
	// required: true
	EventId uuid.UUID `json:"eventId"`
	// VK ID Автора отметки
	// required: false
	CheckAuthorVkId *int64 `json:"checkAuthorVkId"`
	// Время отметки
	// required: true
	Timestamp time.Time `json:"timestamp"`
	// Тип отметки
	// required: true
	CheckType string `json:"checkType"`
	// Широта
	// required: false
	Lat *float64 `json:"lat"`
	// Долгота
	// required: false
	Long *float64 `json:"long"`
}
