package dto

import (
	"time"

	"github.com/google/uuid"
)

// swagger:model EventMiniDto
type EventMiniDto struct {
	// ID События
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id"`
	// Название События
	// required: true
	Title string `json:"title"`
	// Организаторы События
	// required: true
	// default: []
	Orgs []UserMiniDto `json:"orgs"`
	// Участники События
	// required: true
	// default: []
	Participants []UserMiniDto `json:"participants"`
	// Количество участников События
	// required: true
	// min: 0
	ParticipantsCount int `json:"participantsCount"`
	// Статус события
	// required: true
	Status string `json:"status"`
	// Время начала События
	// required: true
	StartsAt time.Time `json:"startAt"`
}

// swagger:model EventDto
type EventDto struct {
	// ID События
	// format: uuid
	// required: true
	ID uuid.UUID `json:"id"`
	// Ссылка на пост ВК
	// required: false
	VkPostLink *string `json:"vkPostLink"`
	// Ссылка на фото
	// required: false
	PhotoURL *string `json:"photoUrl"`

	// Название События
	// required: true
	Title string `json:"title"`
	// Описание События
	// required: true
	Description *string `json:"description"`
	// Вложения События
	// required: true
	// default: []
	Attachments []string `json:"attachments"`
	// Организаторы События
	// required: true
	// default: []
	Orgs []UserMiniDto `json:"orgs"`
	// Адрес проведения События
	// required: false
	Address *string `json:"address"`

	// Участники События
	// required: true
	// default: []
	Participants []UserMiniDto `json:"participants"`
	// Количество участников События
	// required: true
	// min: 0
	ParticipantsCount int `json:"participantsCount"`
	// Статус события
	// required: true
	Status string `json:"status"`
	// Время начала События
	// required: true
	StartsAt time.Time `json:"startAt"`
	// Текущий пользователь является участником
	// required: false
	// default: false
	CurrentUserIsParticipant *bool `json:"currentUserIsParticipant,omitempty"`
}

// swagger:model CreateUpdateEventDto
type CreateUpdateEventDto struct {
	// Ссылка на пост ВК
	// required: false
	VkPostLink *string `json:"vkPostLink"`
	// Ссылка на фото
	// required: false
	PhotoURL *string `json:"photoUrl"`

	// Название События
	// required: false
	Title *string `json:"title"`
	// Описание События
	// required: false
	Description *string `json:"description"`
	// Вложения События
	// required: false
	// default: []
	Attachments *[]string `json:"attachments"`
	// Адрес проведения События
	// required: false
	Address *string `json:"address"`
	// Широта проведения События
	// required: false
	Lat *float64 `json:"lat"`
	// Долгота проведения События
	// required: false
	Long *float64 `json:"long"`
	// Организаторы События
	// required: false
	// default: []
	Orgs []uuid.UUID `json:"orgs"`
	// Статус события
	// required: false
	Status string `json:"status"`
	// Время начала События
	// required: false
	StartsAt *time.Time `json:"startsAt"`
	// ID Доступных ролей для участников
	// required: false
	// default: []
	AvailableRoles []uuid.UUID `json:"availableRoles"` // TODO передать на коды ролей
}
