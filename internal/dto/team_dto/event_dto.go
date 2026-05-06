package team_dto

import (
	"ne_noy/internal/dto"

	"github.com/google/uuid"
)

type TeamDto struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Captain      dto.UserMiniDto   `json:"captain"`
	Members      []dto.UserMiniDto `json:"members"` // 3 участника команды без учета капитана
	TotalMembers int               `json:"total_members"`
}

type CreateTeamDto struct {
	Name      string          `json:"name"`
	CaptainID dto.UserMiniDto `json:"captain_id"`
}
