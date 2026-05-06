package as_team

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

type Team struct {
	model.BaseModel
	CaptainID uuid.UUID
	Captain   model.User
	EventID   uuid.UUID
	Event     AsTeam
	TeamName  string

	Members []TeamMember
}

func (t Team) TableName() string {
	return "teams"
}

type TeamMember struct {
	model.BaseModel
	TeamID uuid.UUID
	Team   Team
	UserID uuid.UUID
	User   model.User
}

func (t TeamMember) TableName() string {
	return "team_members"
}
