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
	ParticipantsCount int           `json:"participants_count"`
	StartsAt          time.Time     `json:"start_at"`
}

type EventDto struct {
	ID         uuid.UUID `json:"id"`
	VkPostLink string    `json:"vk_post_link"`
	PhotoURL   string    `json:"photo_url"`

	Title       string        `json:"title"`
	Description string        `json:"description"`
	Attachments []string      `json:"attachments"`
	Orgs        []UserMiniDto `json:"orgs"`
	Place       string        `json:"place"`

	Participants             []UserMiniDto `json:"participants"`
	ParticipantsCount        int           `json:"participants_count"`
	StartDate                time.Time     `json:"start_date"`
	CurrentUserIsParticipant *bool         `json:"current_user_is_participant,omitempty"`
}
