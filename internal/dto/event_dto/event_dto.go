package event_dto

import (
	"ne_noy/internal/dto"
	"time"

	"github.com/google/uuid"
)

type EventDto struct {
	ID                       uuid.UUID           `json:"id"`
	VkPostId                 *int64              `json:"vkPostId"`
	VkVoteID                 *int64              `json:"vk_vote_id"`
	VkPollAnswerID           *int64              `json:"vk_poll_answer_id"`
	Lat                      *float64            `json:"lat"`
	Long                     *float64            `json:"long"`
	PhotoURL                 *string             `json:"photoUrl"`
	Title                    string              `json:"title"`
	Description              *string             `json:"description"`
	Attachments              []dto.AttachmentDto `json:"attachments"`
	Orgs                     []dto.UserMiniDto   `json:"orgs"`
	Address                  *string             `json:"address"`
	AdAddress                *string             `json:"adAddress"`
	Participants             []dto.UserMiniDto   `json:"participants"`
	ParticipantsCount        int                 `json:"participantsCount"`
	Status                   string              `json:"status"`
	StartsAt                 time.Time           `json:"startAt"`
	EndsAt                   time.Time           `json:"endsAt"`
	CurrentUserIsParticipant *bool               `json:"currentUserIsParticipant,omitempty"`
}

type CreateUpdateEventDto struct {
	VkPostId       *int64  `json:"vkPostId"`
	VkVoteID       *int64  `json:"vk_vote_id"`
	VkPollAnswerID *int64  `json:"vk_poll_answer_id"`
	PhotoURL       *string `json:"photoUrl"`

	Title          *string           `json:"title"`
	Description    *string           `json:"description"`
	Attachments    *[]string         `json:"attachments"`
	Address        *string           `json:"address"`
	AdAddress      *string           `json:"adAddress"`
	Lat            *float64          `json:"lat"`
	Long           *float64          `json:"long"`
	Orgs           []dto.UserMiniDto `json:"orgs"`
	Status         string            `json:"status"`
	StartsAt       *time.Time        `json:"startsAt"`
	EndsAt         *time.Time        `json:"endsAt"`
	AvailableRoles []string          `json:"availableRoles"`
}
