package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_team"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEventTeamRepositoryCreateJoinAndLeave(t *testing.T) {
	ctx := context.Background()
	pool := setupTeamRepositoryPostgres(t)
	repo := NewEventTeamRepository(pool)

	eventID := uuid.New()
	captainID := seedTeamRepoUser(t, ctx, pool, 1001, "Ivan", "Captain")
	memberID := seedTeamRepoUser(t, ctx, pool, 1002, "Petr", "Member")
	seedTeamRepoEvent(t, ctx, pool, eventID, 2, nil)

	created, err := repo.CreateTeam(ctx, eventID, captainID, "Red")
	require.NoError(t, err)
	require.Equal(t, "Red", created.TeamName)
	require.Equal(t, captainID, created.CaptainID)
	require.Equal(t, int64(1001), created.Captain.VkID)
	require.Empty(t, created.Members)

	require.NoError(t, repo.AddMember(ctx, created.ID, memberID))
	require.NoError(t, repo.AddMember(ctx, created.ID, memberID))

	teams, err := repo.GetTeamsByEvent(ctx, eventID)
	require.NoError(t, err)
	require.Len(t, teams, 1)
	require.Len(t, teams[0].Members, 1)
	require.Equal(t, memberID, teams[0].Members[0].UserID)
	require.Equal(t, int64(1002), teams[0].Members[0].User.VkID)

	require.NoError(t, repo.RemoveMember(ctx, created.ID, memberID))

	withoutMember, err := repo.GetTeamByID(ctx, created.ID)
	require.NoError(t, err)
	require.Empty(t, withoutMember.Members)

	err = repo.RemoveMember(ctx, created.ID, memberID)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestEventTeamRepositoryGetEventByID(t *testing.T) {
	ctx := context.Background()
	pool := setupTeamRepositoryPostgres(t)
	repo := NewEventTeamRepository(pool)

	eventID := uuid.New()
	capMax := 4
	seedTeamRepoEvent(t, ctx, pool, eventID, 3, &capMax)

	event, err := repo.GetEventByID(ctx, eventID)
	require.NoError(t, err)
	require.Equal(t, eventID, event.ID)
	require.Equal(t, 3, event.TeamsConstraint)
	require.NotNil(t, event.TeamsCapMax)
	require.Equal(t, 4, *event.TeamsCapMax)
}

func TestEventTeamRepositoryCreateUpdateDeleteEvent(t *testing.T) {
	ctx := context.Background()
	pool := setupTeamRepositoryPostgres(t)
	repo := NewEventTeamRepository(pool)

	startsAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(2 * time.Hour)
	description := "Team event description"
	cover := "https://example.com/team-cover.jpg"
	address := "Moscow"
	additionalAddress := "Building 1"
	lat := 55.7558
	lon := 37.6173
	capMin := 2
	capMax := 5
	vkPostID := int64(12345)

	created, err := repo.CreateEvent(ctx, &as_team.AsTeam{
		EventProfile: events.EventProfile{
			Name:        "Hackathon",
			Description: &description,
			Cover:       &cover,
			Status:      "draft",
			StartsAt:    startsAt,
			EndsAt:      &endsAt,
		},
		TeamsConstraint:   8,
		TeamsCapMin:       &capMin,
		TeamsCapMax:       &capMax,
		Lat:               &lat,
		Lon:               &lon,
		Address:           &address,
		AdditionalAddress: &additionalAddress,
		VkPostID:          &vkPostID,
	})
	require.NoError(t, err)
	require.Equal(t, "Hackathon", created.Name)
	require.Equal(t, 8, created.TeamsConstraint)
	require.NotEqual(t, uuid.Nil, created.ID)

	updatedName := "Updated Hackathon"
	updatedStatus := "active"
	updatedCapMax := 6
	updated, err := repo.UpdateEvent(ctx, created.ID, as_team.AsTeam{
		EventProfile: events.EventProfile{
			Name:   updatedName,
			Status: updatedStatus,
		},
		TeamsCapMax: &updatedCapMax,
	})
	require.NoError(t, err)
	require.Equal(t, updatedName, updated.Name)
	require.Equal(t, updatedStatus, updated.Status)
	require.Equal(t, 8, updated.TeamsConstraint)
	require.NotNil(t, updated.TeamsCapMax)
	require.Equal(t, updatedCapMax, *updated.TeamsCapMax)

	captainID := seedTeamRepoUser(t, ctx, pool, 1003, "Team", "Captain")
	team, err := repo.CreateTeam(ctx, created.ID, captainID, "Backend")
	require.NoError(t, err)

	require.NoError(t, repo.DeleteEvent(ctx, created.ID))

	_, err = repo.GetEventByID(ctx, created.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))

	_, err = repo.GetTeamByID(ctx, team.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func setupTeamRepositoryPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	container, err := tcPostgres.Run(ctx,
		"postgres:17",
		tcPostgres.WithDatabase("team_repo_test"),
		tcPostgres.WithUsername("team_repo_test"),
		tcPostgres.WithPassword("team_repo_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(45*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	sqlDB, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(sqlDB, findTeamRepoMigrationsPath(t)))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))
	t.Cleanup(pool.Close)

	return pool
}

func seedTeamRepoUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, vkID int64, firstName, lastName string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, vk_id, first_name, last_name, photo_url)
		VALUES ($1, $2, $3, $4, $5)
	`, id, vkID, firstName, lastName, "")
	require.NoError(t, err)

	return id
}

func seedTeamRepoEvent(t *testing.T, ctx context.Context, pool *pgxpool.Pool, eventID uuid.UUID, teamsConstraint int, teamsCapMax *int) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO event_as_teams (id, name, status, starts_at, teams_constraint, teams_cap_max)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, eventID, "Team Event", "active", time.Now().UTC(), teamsConstraint, teamsCapMax)
	require.NoError(t, err)
}

func findTeamRepoMigrationsPath(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	cur := wd
	for i := 0; i < 6; i++ {
		path := filepath.Join(cur, "migrations")
		if stat, statErr := os.Stat(path); statErr == nil && stat.IsDir() {
			return path
		}
		cur = filepath.Join(cur, "..")
	}

	require.FailNow(t, fmt.Sprintf("migrations directory not found from %s", wd))
	return ""
}
