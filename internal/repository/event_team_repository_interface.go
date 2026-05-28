package repository

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events/as_team"

	"github.com/google/uuid"
)

type EventTeamRepository interface {
	// GetEventByID возвращает командное мероприятие вместе с настройками вместимости команд.
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*as_team.AsTeam, error)

	// CreateEvent создает запись командного мероприятия в профильной таблице event_as_teams.
	CreateEvent(ctx context.Context, event *as_team.AsTeam) (*as_team.AsTeam, error)

	// UpdateEvent обновляет поля записи командного мероприятия, которые передал сервис.
	UpdateEvent(ctx context.Context, eventID uuid.UUID, update as_team.AsTeam) (*as_team.AsTeam, error)

	// DeleteEvent удаляет запись командного мероприятия вместе с созданными командами и участниками команд.
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error

	// GetTeamsByEvent возвращает команды мероприятия с капитаном и участниками.
	GetTeamsByEvent(ctx context.Context, eventID uuid.UUID) ([]as_team.Team, error)

	// GetTeamByID возвращает одну команду с капитаном и участниками.
	GetTeamByID(ctx context.Context, teamID uuid.UUID) (*as_team.Team, error)

	// CreateTeam создает команду в мероприятии и назначает капитана.
	CreateTeam(ctx context.Context, eventID, captainID uuid.UUID, name string) (*as_team.Team, error)

	// DeleteTeam удаляет команду в мероприятии
	DeleteTeam(ctx context.Context, eventID uuid.UUID) error

	// AddMember добавляет пользователя в команду.
	AddMember(ctx context.Context, teamID, userID uuid.UUID) error

	// RemoveMember удаляет пользователя из команды.
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error

	// UpdateCaptain назначает нового капитана команды; старый капитан становится рядовым участником.
	UpdateCaptain(ctx context.Context, teamID, newCaptainID uuid.UUID) error

	// SetEventOrganizers заменяет список организаторов командного мероприятия.
	SetEventOrganizers(ctx context.Context, eventID uuid.UUID, userIDs []uuid.UUID) error

	// GetEventOrganizers возвращает список организаторов командного мероприятия.
	GetEventOrganizers(ctx context.Context, eventID uuid.UUID) ([]model.User, error)
}
