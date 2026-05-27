package event_as_team

import (
	"context"
	"errors"
	"ne_noy/internal/dto"
	"ne_noy/internal/dto/team_dto"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
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

func TestEventTeamServiceCreateUpdateDeleteTeamEvent(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEventTeamRepo()
	service := NewEventTeamService(repo, &fakeVkClient{})

	startsAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	capMin := 2
	capMax := 5
	created, err := service.CreateTeamEvent(ctx, team_dto.CreateTeamEventDto{
		Name:            "Team Event",
		Status:          "draft",
		StartsAt:        dto.FlexTime{Time: startsAt},
		TeamsConstraint: 4,
		TeamsCapMin:     &capMin,
		TeamsCapMax:     &capMax,
	})
	require.NoError(t, err)
	require.Equal(t, "Team Event", created.Name)
	require.Equal(t, 4, created.TeamsConstraint)
	require.NotEqual(t, uuid.Nil, created.ID)

	updatedName := "Updated Team Event"
	updatedCapMax := 6
	updated, err := service.UpdateTeamEvent(ctx, created.ID, team_dto.UpdateTeamEventDto{
		Name:        &updatedName,
		TeamsCapMax: &updatedCapMax,
	})
	require.NoError(t, err)
	require.Equal(t, updatedName, updated.Name)
	require.Equal(t, 4, updated.TeamsConstraint)
	require.NotNil(t, updated.TeamsCapMax)
	require.Equal(t, updatedCapMax, *updated.TeamsCapMax)

	require.NoError(t, service.DeleteTeamEvent(ctx, team_dto.DeleteTeamEventDto{ID: created.ID}))
	_, err = service.GetTeamEvent(ctx, created.ID)
	require.Error(t, err)
}

func TestEventTeamServiceCreateTeamEventValidatesRequiredFields(t *testing.T) {
	service := NewEventTeamService(newFakeEventTeamRepo(), &fakeVkClient{})

	_, err := service.CreateTeamEvent(context.Background(), team_dto.CreateTeamEventDto{
		Status:   "draft",
		StartsAt: dto.FlexTime{Time: time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "team event name is required")
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

func TestEventTeamServiceCreateTeamEventWithOrganizers(t *testing.T) {
	ctx := context.Background()
	orgID1 := uuid.New()
	orgID2 := uuid.New()
	repo := newFakeEventTeamRepo()
	repo.users[orgID1] = model.User{BaseModel: model.BaseModel{ID: orgID1}, FirstName: "Org1"}
	repo.users[orgID2] = model.User{BaseModel: model.BaseModel{ID: orgID2}, FirstName: "Org2"}
	service := NewEventTeamService(repo, &fakeVkClient{})

	created, err := service.CreateTeamEvent(ctx, team_dto.CreateTeamEventDto{
		Name:            "Orgs Event",
		Status:          "draft",
		StartsAt:        dto.FlexTime{Time: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)},
		TeamsConstraint: 2,
		Organizers:      []uuid.UUID{orgID1, orgID2},
	})
	require.NoError(t, err)
	require.Len(t, created.Organizers, 2)
}

func TestEventTeamServiceGetTeamEventIncludesOrganizers(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.New()
	orgID := uuid.New()
	repo := newFakeEventTeamRepo()
	repo.events[eventID] = &as_team.AsTeam{}
	repo.organizers[eventID] = []model.User{
		{BaseModel: model.BaseModel{ID: orgID}, FirstName: "Org", VkID: 3001},
	}
	service := NewEventTeamService(repo, &fakeVkClient{})

	result, err := service.GetTeamEvent(ctx, eventID)
	require.NoError(t, err)
	require.Len(t, result.Organizers, 1)
	require.Equal(t, int64(3001), result.Organizers[0].VkId)
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
	events     map[uuid.UUID]*as_team.AsTeam
	teams      map[uuid.UUID]as_team.Team
	users      map[uuid.UUID]model.User
	organizers map[uuid.UUID][]model.User
}

func newFakeEventTeamRepo() *fakeEventTeamRepo {
	return &fakeEventTeamRepo{
		events:     make(map[uuid.UUID]*as_team.AsTeam),
		teams:      make(map[uuid.UUID]as_team.Team),
		users:      make(map[uuid.UUID]model.User),
		organizers: make(map[uuid.UUID][]model.User),
	}
}

func (f *fakeEventTeamRepo) GetEventByID(_ context.Context, eventID uuid.UUID) (*as_team.AsTeam, error) {
	event, ok := f.events[eventID]
	if !ok {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (f *fakeEventTeamRepo) CreateEvent(_ context.Context, event *as_team.AsTeam) (*as_team.AsTeam, error) {
	eventID := uuid.New()
	created := *event
	created.EventProfile = events.EventProfile{
		BaseModel:   model.BaseModel{ID: eventID},
		Name:        event.Name,
		Description: event.Description,
		Cover:       event.Cover,
		Status:      event.Status,
		StartsAt:    event.StartsAt,
		EndsAt:      event.EndsAt,
	}
	f.events[eventID] = &created
	return &created, nil
}

func (f *fakeEventTeamRepo) UpdateEvent(_ context.Context, eventID uuid.UUID, update as_team.AsTeam) (*as_team.AsTeam, error) {
	event, ok := f.events[eventID]
	if !ok {
		return nil, errors.New("event not found")
	}
	if update.Name != "" {
		event.Name = update.Name
	}
	if update.Description != nil {
		event.Description = update.Description
	}
	if update.Cover != nil {
		event.Cover = update.Cover
	}
	if update.Status != "" {
		event.Status = update.Status
	}
	if !update.StartsAt.IsZero() {
		event.StartsAt = update.StartsAt
	}
	if update.EndsAt != nil {
		event.EndsAt = update.EndsAt
	}
	if update.TeamsConstraint != 0 {
		event.TeamsConstraint = update.TeamsConstraint
	}
	if update.TeamsCapMin != nil {
		event.TeamsCapMin = update.TeamsCapMin
	}
	if update.TeamsCapMax != nil {
		event.TeamsCapMax = update.TeamsCapMax
	}
	if update.Lat != nil {
		event.Lat = update.Lat
	}
	if update.Lon != nil {
		event.Lon = update.Lon
	}
	if update.Address != nil {
		event.Address = update.Address
	}
	if update.AdditionalAddress != nil {
		event.AdditionalAddress = update.AdditionalAddress
	}
	if update.VkPostID != nil {
		event.VkPostID = update.VkPostID
	}
	return event, nil
}

func (f *fakeEventTeamRepo) DeleteEvent(_ context.Context, eventID uuid.UUID) error {
	if _, ok := f.events[eventID]; !ok {
		return errors.New("event not found")
	}
	delete(f.events, eventID)
	for teamID, team := range f.teams {
		if team.EventID == eventID {
			delete(f.teams, teamID)
		}
	}
	return nil
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

func (f *fakeEventTeamRepo) SetEventOrganizers(_ context.Context, eventID uuid.UUID, userIDs []uuid.UUID) error {
	orgs := make([]model.User, 0, len(userIDs))
	for _, id := range userIDs {
		if u, ok := f.users[id]; ok {
			orgs = append(orgs, u)
		} else {
			orgs = append(orgs, model.User{BaseModel: model.BaseModel{ID: id}})
		}
	}
	f.organizers[eventID] = orgs
	return nil
}

func (f *fakeEventTeamRepo) GetEventOrganizers(_ context.Context, eventID uuid.UUID) ([]model.User, error) {
	return f.organizers[eventID], nil
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
