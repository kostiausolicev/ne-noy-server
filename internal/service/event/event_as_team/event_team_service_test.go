package event_as_team

import (
	"context"
	"errors"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/team_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events/as_team"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEventTeamServiceCreateTeamChecksTeamLimit(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.New()
	captainID := uuid.New()

	repo := newFakeEventTeamRepo()
	repo.events[eventID] = &as_team.AsTeam{TeamsConstraint: 1}
	repo.teams[uuid.New()] = as_team.Team{EventID: eventID}

	service := NewEventTeamService(repo, &fakeVkClient{})

	_, err := service.CreateTeam(ctx, eventID, team_dto.CreateTeamDto{
		Name:      "Red",
		CaptainID: captainID,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "teams limit reached")
}

func TestEventTeamServiceJoinTeamChecksCapacityAndDuplicate(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.New()
	teamID := uuid.New()
	captainID := uuid.New()
	memberID := uuid.New()
	capMax := 2

	repo := newFakeEventTeamRepo()
	repo.events[eventID] = &as_team.AsTeam{TeamsConstraint: 2, TeamsCapMax: &capMax}
	repo.users[memberID] = model.User{BaseModel: model.BaseModel{ID: memberID}, VkID: 2002, FirstName: "Petr"}
	repo.teams[teamID] = as_team.Team{
		BaseModel: model.BaseModel{ID: teamID},
		EventID:   eventID,
		CaptainID: captainID,
		Captain:   model.User{BaseModel: model.BaseModel{ID: captainID}, VkID: 2001, FirstName: "Ivan"},
	}

	service := NewEventTeamService(repo, &fakeVkClient{})

	require.NoError(t, service.JoinTeam(ctx, teamID, memberID))
	require.Len(t, repo.teams[teamID].Members, 1)

	require.NoError(t, service.JoinTeam(ctx, teamID, memberID))
	require.Len(t, repo.teams[teamID].Members, 1)

	anotherID := uuid.New()
	err := service.JoinTeam(ctx, teamID, anotherID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "team capacity reached")
}

func TestEventTeamServiceGetTeamsMapsDtoAndLimitsPreviewMembers(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.New()
	teamID := uuid.New()
	captainID := uuid.New()

	repo := newFakeEventTeamRepo()
	repo.teams[teamID] = as_team.Team{
		BaseModel: model.BaseModel{ID: teamID},
		EventID:   eventID,
		TeamName:  "Red",
		CaptainID: captainID,
		Captain:   model.User{BaseModel: model.BaseModel{ID: captainID}, VkID: 2001, FirstName: "Ivan"},
		Members: []as_team.TeamMember{
			fakeTeamMember(2002),
			fakeTeamMember(2003),
			fakeTeamMember(2004),
			fakeTeamMember(2005),
		},
	}

	service := NewEventTeamService(repo, &fakeVkClient{})

	teams, err := service.GetTeamsOnEvent(ctx, eventID)
	require.NoError(t, err)
	require.Len(t, teams, 1)
	require.Len(t, teams[0].Members, 3)
	require.Equal(t, 5, teams[0].TotalMembers)
	require.Equal(t, int64(2001), teams[0].Captain.VkId)
}

func TestEventTeamServiceSendNotificationToTeam(t *testing.T) {
	ctx := context.Background()
	teamID := uuid.New()
	captainID := uuid.New()

	repo := newFakeEventTeamRepo()
	repo.teams[teamID] = as_team.Team{
		BaseModel: model.BaseModel{ID: teamID},
		CaptainID: captainID,
		Captain:   model.User{BaseModel: model.BaseModel{ID: captainID}, VkID: 2001},
		Members: []as_team.TeamMember{
			{User: model.User{VkID: 2002}},
			{User: model.User{VkID: 2003}},
		},
	}
	client := &fakeVkClient{}
	service := NewEventTeamService(repo, client)

	require.NoError(t, service.SendNotificationToTeam(ctx, teamID, "hello"))
	require.Equal(t, []string{"2001", "2002", "2003"}, client.userIDs)
	require.Equal(t, "hello", client.message)
}

type fakeEventTeamRepo struct {
	events map[uuid.UUID]*as_team.AsTeam
	teams  map[uuid.UUID]as_team.Team
	users  map[uuid.UUID]model.User
}

func newFakeEventTeamRepo() *fakeEventTeamRepo {
	return &fakeEventTeamRepo{
		events: make(map[uuid.UUID]*as_team.AsTeam),
		teams:  make(map[uuid.UUID]as_team.Team),
		users:  make(map[uuid.UUID]model.User),
	}
}

func (f *fakeEventTeamRepo) GetEventByID(_ context.Context, eventID uuid.UUID) (*as_team.AsTeam, error) {
	event, ok := f.events[eventID]
	if !ok {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (f *fakeEventTeamRepo) GetTeamsByEvent(_ context.Context, eventID uuid.UUID) ([]as_team.Team, error) {
	teams := make([]as_team.Team, 0)
	for _, team := range f.teams {
		if team.EventID == eventID {
			teams = append(teams, team)
		}
	}
	return teams, nil
}

func (f *fakeEventTeamRepo) GetTeamByID(_ context.Context, teamID uuid.UUID) (*as_team.Team, error) {
	team, ok := f.teams[teamID]
	if !ok {
		return nil, errors.New("team not found")
	}
	return &team, nil
}

func (f *fakeEventTeamRepo) CreateTeam(_ context.Context, eventID, captainID uuid.UUID, name string) (*as_team.Team, error) {
	teamID := uuid.New()
	team := as_team.Team{
		BaseModel: model.BaseModel{ID: teamID},
		EventID:   eventID,
		TeamName:  name,
		CaptainID: captainID,
		Captain:   model.User{BaseModel: model.BaseModel{ID: captainID}, FirstName: "Captain"},
	}
	f.teams[teamID] = team
	return &team, nil
}

func (f *fakeEventTeamRepo) AddMember(_ context.Context, teamID, userID uuid.UUID) error {
	team := f.teams[teamID]
	user := f.users[userID]
	if user.ID == uuid.Nil {
		user = model.User{BaseModel: model.BaseModel{ID: userID}}
	}
	team.Members = append(team.Members, as_team.TeamMember{
		BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: time.Now().UTC()},
		TeamID:    teamID,
		UserID:    userID,
		User:      user,
	})
	f.teams[teamID] = team
	return nil
}

func (f *fakeEventTeamRepo) RemoveMember(_ context.Context, teamID, userID uuid.UUID) error {
	team := f.teams[teamID]
	for i, member := range team.Members {
		if member.UserID == userID {
			team.Members = append(team.Members[:i], team.Members[i+1:]...)
			f.teams[teamID] = team
			return nil
		}
	}
	return errors.New("member not found")
}

type fakeVkClient struct {
	userIDs []string
	message string
}

func (f *fakeVkClient) GetVkUsers(_ []string) ([]dto.CreateUserDto, error) {
	return nil, nil
}

func (f *fakeVkClient) SendNotification(userIDs []string, messageText, _ string) (dto.SendMessageResponse, error) {
	f.userIDs = append([]string(nil), userIDs...)
	f.message = messageText
	return dto.SendMessageResponse{}, nil
}

func (f *fakeVkClient) CreateChat(_ context.Context, _ string, _ []uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeVkClient) AddUserToChat(_ context.Context, _ string, _, _ int64) error {
	return nil
}

func fakeTeamMember(vkID int64) as_team.TeamMember {
	userID := uuid.New()
	return as_team.TeamMember{
		BaseModel: model.BaseModel{ID: uuid.New()},
		UserID:    userID,
		User: model.User{
			BaseModel: model.BaseModel{ID: userID},
			VkID:      vkID,
		},
	}
}
