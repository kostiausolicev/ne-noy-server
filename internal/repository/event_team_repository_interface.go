package repository

import (
	"context"
	"ne_noy/internal/model/events/as_team"

	"github.com/google/uuid"
)

type EventTeamRepository interface {
	// GetEventByID возвращает командное мероприятие вместе с настройками вместимости команд.
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*as_team.AsTeam, error)

	// GetTeamsByEvent возвращает команды мероприятия с капитаном и участниками.
	GetTeamsByEvent(ctx context.Context, eventID uuid.UUID) ([]as_team.Team, error)

	// GetTeamByID возвращает одну команду с капитаном и участниками.
	GetTeamByID(ctx context.Context, teamID uuid.UUID) (*as_team.Team, error)

	// CreateTeam создает команду в мероприятии и назначает капитана.
	CreateTeam(ctx context.Context, eventID, captainID uuid.UUID, name string) (*as_team.Team, error)

	// AddMember добавляет пользователя в команду.
	AddMember(ctx context.Context, teamID, userID uuid.UUID) error

	// RemoveMember удаляет пользователя из команды.
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
}
