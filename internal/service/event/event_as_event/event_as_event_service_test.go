package event_as_event

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ne_noy/internal/dto"
	"ne_noy/internal/dto/event_dto"
	"ne_noy/internal/repository/impl"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEventAsEventServiceCreateGetAndUpdateReturnsEventDto(t *testing.T) {
	ctx := context.Background()
	pool := setupEventAsEventServicePostgres(t)
	service := NewEventAsEventService(impl.NewEventEventRepository(pool))

	orgID := uuid.MustParse("11111111-aaaa-4aaa-8aaa-111111111111")
	memberID := uuid.MustParse("22222222-bbbb-4bbb-8bbb-222222222222")
	seedEventAsEventServiceUser(t, ctx, pool, orgID, 1001, "Ivan", "Org", "https://example.com/org.jpg")
	seedEventAsEventServiceUser(t, ctx, pool, memberID, 1002, "Petr", "Member", "https://example.com/member.jpg")

	description := "Initial description"
	cover := "https://example.com/cover.jpg"
	address := "Main street"
	additionalAddress := "Gate 1"
	lat := 55.751244
	lon := 37.618423
	vkPostID := int64(5001)
	vkVoteID := int64(6001)
	vkPollAnswerID := int64(7001)
	startsAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	endsAt := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	title := "Service event"

	created, err := service.CreateEvent(ctx, event_dto.CreateUpdateEventDto{
		VkPostId:       &vkPostID,
		VkVoteID:       &vkVoteID,
		VkPollAnswerID: &vkPollAnswerID,
		PhotoURL:       &cover,
		Title:          &title,
		Description:    &description,
		Address:        &address,
		AdAddress:      &additionalAddress,
		Lat:            &lat,
		Long:           &lon,
		Orgs: []dto.UserMiniDto{
			{
				ID:        orgID,
				VkId:      1001,
				FirstName: "Ivan",
				LastName:  "Org",
				PhotoURL:  "https://example.com/org.jpg",
			},
		},
		Status:   "active",
		StartsAt: &startsAt,
		EndsAt:   &endsAt,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)
	require.Equal(t, event_dto.EventDto{
		ID:             created.ID,
		VkPostId:       &vkPostID,
		VkVoteID:       &vkVoteID,
		VkPollAnswerID: &vkPollAnswerID,
		Lat:            &lat,
		Long:           &lon,
		PhotoURL:       &cover,
		Title:          "Service event",
		Description:    &description,
		Attachments:    []dto.AttachmentDto{},
		Orgs: []dto.UserMiniDto{
			{
				ID:        orgID,
				VkId:      1001,
				FirstName: "Ivan",
				LastName:  "Org",
				PhotoURL:  "https://example.com/org.jpg",
			},
		},
		Address:           &address,
		AdAddress:         &additionalAddress,
		Participants:      []dto.UserMiniDto{},
		ParticipantsCount: 0,
		Status:            "active",
		StartsAt:          startsAt,
		EndsAt:            endsAt,
	}, created)

	seedEventAsEventServiceParticipant(t, ctx, pool, created.ID, memberID)

	got, err := service.GetEventById(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, 1, got.ParticipantsCount)
	require.Equal(t, []dto.UserMiniDto{
		{
			ID:        memberID,
			VkId:      1002,
			FirstName: "Petr",
			LastName:  "Member",
			PhotoURL:  "https://example.com/member.jpg",
		},
	}, got.Participants)

	updatedTitle := "Updated service event"
	updatedStatus := "done"
	updatedDescription := "Updated description"
	updated, err := service.UpdateEvent(ctx, created.ID, event_dto.CreateUpdateEventDto{
		Title:       &updatedTitle,
		Description: &updatedDescription,
		Status:      updatedStatus,
	})
	require.NoError(t, err)
	require.Equal(t, updatedTitle, updated.Title)
	require.Equal(t, updatedStatus, updated.Status)
	require.Equal(t, &updatedDescription, updated.Description)
	require.Equal(t, 1, updated.ParticipantsCount)
	require.Equal(t, got.Participants, updated.Participants)
}

func TestEventAsEventServiceGetEventParticipantsReturnsUserMiniDtos(t *testing.T) {
	ctx := context.Background()
	pool := setupEventAsEventServicePostgres(t)
	service := NewEventAsEventService(impl.NewEventEventRepository(pool))

	eventID := uuid.MustParse("33333333-cccc-4ccc-8ccc-333333333333")
	memberID := uuid.MustParse("44444444-dddd-4ddd-8ddd-444444444444")
	seedEventAsEventServiceEvent(t, ctx, pool, eventID)
	seedEventAsEventServiceUser(t, ctx, pool, memberID, 1003, "Anna", "Participant", "https://example.com/participant.jpg")
	seedEventAsEventServiceParticipant(t, ctx, pool, eventID, memberID)

	participants, err := service.GetEventParticipants(ctx, eventID)

	require.NoError(t, err)
	require.Equal(t, []dto.UserMiniDto{
		{
			ID:        memberID,
			VkId:      1003,
			FirstName: "Anna",
			LastName:  "Participant",
			PhotoURL:  "https://example.com/participant.jpg",
		},
	}, participants)
}

func setupEventAsEventServicePostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	container, err := tcPostgres.Run(ctx,
		"postgres:17",
		tcPostgres.WithDatabase("event_as_event_service_test"),
		tcPostgres.WithUsername("event_as_event_service_test"),
		tcPostgres.WithPassword("event_as_event_service_test"),
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
	require.NoError(t, goose.Up(sqlDB, findEventAsEventServiceMigrationsPath(t)))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))
	t.Cleanup(pool.Close)

	return pool
}

func seedEventAsEventServiceUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, vkID int64, firstName, lastName, photoURL string) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, vk_id, first_name, last_name, role_id, photo_url, geo_available, notification_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, vkID, firstName, lastName, uuid.MustParse("9ef01b95-6d87-4115-80df-7085a647bf36"), photoURL, false, true)
	require.NoError(t, err)
}

func seedEventAsEventServiceEvent(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO event_as_events (id, name, status, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, "Seeded event", "active", time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC), time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
}

func seedEventAsEventServiceParticipant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, eventID, userID uuid.UUID) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO event_participants (event_id, user_id, prepare_type)
		VALUES ($1, $2, $3)
	`, eventID, userID, "app")
	require.NoError(t, err)
}

func findEventAsEventServiceMigrationsPath(t *testing.T) string {
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
