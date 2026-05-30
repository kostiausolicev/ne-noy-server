package test_dto

import (
	"ne_noy/internal/dto"
	"time"

	"github.com/google/uuid"
)

type TestDto struct {
	ID             uuid.UUID           `json:"id"`
	Name           string              `json:"name"`
	Description    *string             `json:"description"`
	Cover          *string             `json:"cover"`
	Status         string              `json:"status"`
	StartsAt       time.Time           `json:"starts_at"`
	EndsAt         *time.Time          `json:"ends_at"`
	ExtLinkID      *string             `json:"ext_link_id"`
	Attempts       int                 `json:"attempts"`
	VkPostID       *int64              `json:"vk_post_id"`
	AvailableRoles []string            `json:"available_roles"`
	Organizers     []dto.UserMiniDto   `json:"organizers"`
	Attachments    []dto.AttachmentDto `json:"attachments"`
	Questions      []QuestionDto       `json:"questions"`
}

type CreateTestDto struct {
	Name           string              `json:"name"`
	Description    *string             `json:"description"`
	Cover          *string             `json:"cover"`
	Status         string              `json:"status"`
	StartsAt       dto.FlexTime        `json:"starts_at"`
	EndsAt         *dto.FlexTime       `json:"ends_at"`
	ExtLinkID      *string             `json:"ext_link_id"`
	Attempts       int                 `json:"attempts"`
	VkPostID       *int64              `json:"vk_post_id"`
	AvailableRoles []string            `json:"available_roles"`
	Organizers     []uuid.UUID         `json:"organizers"`
	Attachments    []dto.AttachmentDto `json:"attachments"`
}

type UpdateTestDto struct {
	Name           *string              `json:"name"`
	Description    *string              `json:"description"`
	Cover          *string              `json:"cover"`
	Status         *string              `json:"status"`
	StartsAt       *dto.FlexTime        `json:"starts_at"`
	EndsAt         *dto.FlexTime        `json:"ends_at"`
	ExtLinkID      *string              `json:"ext_link_id"`
	Attempts       *int                 `json:"attempts"`
	VkPostID       *int64               `json:"vk_post_id"`
	AvailableRoles []string             `json:"available_roles"`
	Organizers     *[]uuid.UUID         `json:"organizers"`
	Attachments    *[]dto.AttachmentDto `json:"attachments"`
}

type DeleteTestDto struct {
	ID uuid.UUID `json:"id"`
}

type TestReportDto struct {
	DownloadURL string `json:"download_url"`
}
