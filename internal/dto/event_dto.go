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
	StartsAt          time.Time     `json:"startAt"`
}

type EventDto struct {
	ID         uuid.UUID `json:"id"`
	VkPostLink *string   `json:"vkPostLink"`
	PhotoURL   *string   `json:"photoUrl"`

	Title       string        `json:"title"`
	Description *string       `json:"description"`
	Attachments []string      `json:"attachments"`
	Orgs        []UserMiniDto `json:"orgs"`
	Address     *string       `json:"address"`

	Participants             []UserMiniDto `json:"participants"`
	ParticipantsCount        int           `json:"participantsCount"`
	StartsAt                 *time.Time    `json:"startsAt"`
	CurrentUserIsParticipant *bool         `json:"currentUserIsParticipant,omitempty"`
}
