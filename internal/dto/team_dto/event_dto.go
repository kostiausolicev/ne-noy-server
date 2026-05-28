package team_dto

import (
	"time"

	"github.com/google/uuid"

	"ne_noy/internal/dto"
)

type TeamEventDto struct {
	ID                uuid.UUID           `json:"id"`
	Name              string              `json:"name"`
	Description       *string             `json:"description"`
	Cover             *string             `json:"cover"`
	Status            string              `json:"status"`
	StartsAt          time.Time           `json:"starts_at"`
	EndsAt            *time.Time          `json:"ends_at"`
	TeamsConstraint   int                 `json:"teams_constraint"`
	TeamsCapMin       *int                `json:"teams_cap_min"`
	TeamsCapMax       *int                `json:"teams_cap_max"`
	Lat               *float64            `json:"lat"`
	Lon               *float64            `json:"lon"`
	Address           *string             `json:"address"`
	AdditionalAddress *string             `json:"additional_address"`
	VkPostID          *int64              `json:"vk_post_id"`
	AvailableRoles    []string            `json:"available_roles"`
	Organizers        []dto.UserMiniDto   `json:"organizers"`
	Attachments       []dto.AttachmentDto `json:"attachments"`
}

type CreateTeamEventDto struct {
	Name              string              `json:"name"`
	Description       *string             `json:"description"`
	Cover             *string             `json:"cover"`
	Status            string              `json:"status"`
	StartsAt          dto.FlexTime        `json:"starts_at"`
	EndsAt            *dto.FlexTime       `json:"ends_at"`
	TeamsConstraint   int                 `json:"teams_constraint"`
	TeamsCapMin       *int                `json:"teams_cap_min"`
	TeamsCapMax       *int                `json:"teams_cap_max"`
	Lat               *float64            `json:"lat"`
	Lon               *float64            `json:"lon"`
	Address           *string             `json:"address"`
	AdditionalAddress *string             `json:"additional_address"`
	VkPostID          *int64              `json:"vk_post_id"`
	AvailableRoles    []string            `json:"available_roles"`
	Organizers        []uuid.UUID         `json:"organizers"`
	Attachments       []dto.AttachmentDto `json:"attachments"`
}

type UpdateTeamEventDto struct {
	Name              *string              `json:"name"`
	Description       *string              `json:"description"`
	Cover             *string              `json:"cover"`
	Status            *string              `json:"status"`
	StartsAt          *dto.FlexTime        `json:"starts_at"`
	EndsAt            *dto.FlexTime        `json:"ends_at"`
	TeamsConstraint   *int                 `json:"teams_constraint"`
	TeamsCapMin       *int                 `json:"teams_cap_min"`
	TeamsCapMax       *int                 `json:"teams_cap_max"`
	Lat               *float64             `json:"lat"`
	Lon               *float64             `json:"lon"`
	Address           *string              `json:"address"`
	AdditionalAddress *string              `json:"additional_address"`
	VkPostID          *int64               `json:"vk_post_id"`
	AvailableRoles    []string             `json:"available_roles"`
	Attachments       *[]dto.AttachmentDto `json:"attachments"`
}

type DeleteTeamEventDto struct {
	ID uuid.UUID `json:"id"`
}

type TeamDto struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Captain      dto.UserMiniDto   `json:"captain"`
	Members      []dto.UserMiniDto `json:"members"` // 3 участника команды без учета капитана
	TotalMembers int               `json:"total_members"`
}

type CreateTeamDto struct {
	Name      string    `json:"name"`
	CaptainID uuid.UUID `json:"captain_id"`
}

type CreateTeamRequestDto struct {
	Name string `json:"name"`
}

type SendTeamNotificationDto struct {
	Message string `json:"message"`
}
