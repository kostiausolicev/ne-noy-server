package event_as_team

import (
	"context"
	"errors"
	"fmt"
	vkClient "ne_noy/internal/client"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/team_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events/as_team"
	"ne_noy/internal/repository"
	"strconv"

	"github.com/google/uuid"
)

type eventTeamService struct {
	repo repository.EventTeamRepository
	cl   vkClient.VkApiClient
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

func NewEventTeamService(repo repository.EventTeamRepository, cl vkClient.VkApiClient) EventTeamService {
	return &eventTeamService{repo: repo, cl: cl}
}

func (e *eventTeamService) GetTeamsOnEvent(ctx context.Context, eventId uuid.UUID) ([]team_dto.TeamDto, error) {
	teams, err := e.repo.GetTeamsByEvent(ctx, eventId)
	if err != nil {
		return nil, err
	}

	result := make([]team_dto.TeamDto, len(teams))
	for i, team := range teams {
		result[i] = e.teamToDto(team)
	}

	return result, nil
}

func (e *eventTeamService) CreateTeam(ctx context.Context, eventId uuid.UUID, team team_dto.CreateTeamDto) (team_dto.TeamDto, error) {
	if team.Name == "" {
		return team_dto.TeamDto{}, errors.New("team name is required")
	}

	event, err := e.repo.GetEventByID(ctx, eventId)
	if err != nil {
		return team_dto.TeamDto{}, err
	}

	teams, err := e.repo.GetTeamsByEvent(ctx, eventId)
	if err != nil {
		return team_dto.TeamDto{}, err
	}
	if event.TeamsConstraint > 0 && len(teams) >= event.TeamsConstraint {
		return team_dto.TeamDto{}, fmt.Errorf("teams limit reached")
	}

	createdTeam, err := e.repo.CreateTeam(ctx, eventId, team.CaptainID, team.Name)
	if err != nil {
		return team_dto.TeamDto{}, err
	}

	return e.teamToDto(*createdTeam), nil
}

func (e *eventTeamService) JoinTeam(ctx context.Context, teamId, userId uuid.UUID) error {
	team, err := e.repo.GetTeamByID(ctx, teamId)
	if err != nil {
		return err
	}

	event, err := e.repo.GetEventByID(ctx, team.EventID)
	if err != nil {
		return err
	}

	// Капитан уже считается участником, поэтому повторное вступление капитана не меняет состав команды.
	if team.CaptainID == userId {
		return nil
	}

	for _, member := range team.Members {
		if member.UserID == userId {
			return nil
		}
	}

	if event.TeamsCapMax != nil && e.totalMembers(*team) >= *event.TeamsCapMax {
		return fmt.Errorf("team capacity reached")
	}

	return e.repo.AddMember(ctx, teamId, userId)
}

func (e *eventTeamService) LeaveTeam(ctx context.Context, teamId, userId uuid.UUID) error {
	team, err := e.repo.GetTeamByID(ctx, teamId)
	if err != nil {
		return err
	}
	if team.CaptainID == userId {
		return errors.New("captain cannot leave team")
	}

	return e.repo.RemoveMember(ctx, teamId, userId)
}

func (e *eventTeamService) SendNotificationToTeam(ctx context.Context, teamId uuid.UUID, message string) error {
	team, err := e.repo.GetTeamByID(ctx, teamId)
	if err != nil {
		return err
	}

	userIDs := make([]string, 0, e.totalMembers(*team))
	userIDs = append(userIDs, strconv.FormatInt(team.Captain.VkID, 10))
	for _, member := range team.Members {
		userIDs = append(userIDs, strconv.FormatInt(member.User.VkID, 10))
	}

	// Фрагмент оставляем пустым: сервис команд сейчас отправляет только текстовое уведомление конкретному составу.
	_, err = e.cl.SendNotification(userIDs, message, "")
	return err
}

func (e *eventTeamService) teamToDto(team as_team.Team) team_dto.TeamDto {
	members := make([]dto.UserMiniDto, 0, min(len(team.Members), 3))
	for i, member := range team.Members {
		if i == 3 {
			break
		}
		members = append(members, userToMiniDto(member.User))
	}

	return team_dto.TeamDto{
		ID:           team.ID,
		Name:         team.TeamName,
		Captain:      userToMiniDto(team.Captain),
		Members:      members,
		TotalMembers: e.totalMembers(team),
	}
}

func (e *eventTeamService) totalMembers(team as_team.Team) int {
	// Капитан хранится отдельно от team_members, но в API считается полноправным участником команды.
	return len(team.Members) + 1
}

func userToMiniDto(user model.User) dto.UserMiniDto {
	return dto.UserMiniDto{
		ID:        user.ID,
		VkId:      user.VkID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		PhotoURL:  user.PhotoURL,
	}
}
