package event_as_team

import (
	"context"
	"ne_noy/internal/dto/team_dto"

	"github.com/google/uuid"
)

type eventTeamService struct {
	// репозиторий
	// VkClient для отправки сообщений
}

type EventTeamService interface {
	// GetTeamsOnEvent возвращает все команды в мероприятии
	GetTeamsOnEvent(ctx context.Context, eventId uuid.UUID) ([]team_dto.TeamDto, error)
	// CreateTeam создает команду в мероприятии
	CreateTeam(ctx context.Context, eventId uuid.UUID, team team_dto.CreateTeamDto) (team_dto.TeamDto, error)
	// JoinTeam присоединение пользователя к команде
	JoinTeam(ctx context.Context, teamId, userId uuid.UUID) error
	// LeaveTeam выход пользователя из команды
	LeaveTeam(ctx context.Context, teamId, userId uuid.UUID) error
	// SendNotificationToTeam отправки уведомлений для команды
	SendNotificationToTeam(ctx context.Context, teamId uuid.UUID, message string) error
}
