package team_dto

import (
	"github.com/google/uuid"

	"ne_noy/internal/dto"
)

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
