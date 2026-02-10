package dto

import (
	"time"

	"github.com/google/uuid"
)

type EventMiniDto struct {
	ID                uuid.UUID     `json:"id"`
	Title             string        `json:"title"`
	Orgs              []UserMiniDto `json:"orgs"`
	Participants      []UserMiniDto `json:"participants"`
	ParticipantsCount int           `json:"participantsCount"`
	Status            string        `json:"status"`
	StartsAt          time.Time     `json:"startAt"`
}

type EventDto struct {
	ID       uuid.UUID `json:"id"`
	VkPostId *int64    `json:"vkPostId"`
	PhotoURL *string   `json:"photoUrl"`

	Title       string        `json:"title"`
	Description *string       `json:"description"`
	Attachments []string      `json:"attachments"`
	Orgs        []UserMiniDto `json:"orgs"`
	Address     *string       `json:"address"`

	Participants             []UserMiniDto `json:"participants"`
	ParticipantsCount        int           `json:"participantsCount"`
	Status                   string        `json:"status"`
	StartsAt                 time.Time     `json:"startAt"`
	CurrentUserIsParticipant *bool         `json:"currentUserIsParticipant,omitempty"`
}

type CreateUpdateEventDto struct {
	VkPostId *int64  `json:"vkPostId"`
	PhotoURL *string `json:"photoUrl"`

	Title             *string     `json:"title"`
	Description       *string     `json:"description"`
	Attachments       *[]string   `json:"attachments"`
	Address           *string     `json:"address"`
	AdditionalAddress *string     `json:"address"`
	Lat               *float64    `json:"lat"`
	Long              *float64    `json:"long"`
	Orgs              []uuid.UUID `json:"orgs"`
	Status            *string     `json:"status"`
	StartsAt          *time.Time  `json:"startsAt"`
	AvailableRoles    []string    `json:"availableRoles"`
}
