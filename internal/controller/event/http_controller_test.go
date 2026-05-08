package event

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"

	vkclient "ne_noy/internal/client"
	"ne_noy/internal/config"
	"ne_noy/internal/controller"
	"ne_noy/internal/controller/middleware"
	"ne_noy/internal/repository/impl"
	appservice "ne_noy/internal/service"
	event_service "ne_noy/internal/service/event"
	event_as_event_service "ne_noy/internal/service/event/event_as_event"
	event_as_team_service "ne_noy/internal/service/event/event_as_team"
	event_as_test_service "ne_noy/internal/service/event/event_as_test"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	httpControllerSecret = "http-controller-secret"
	httpControllerAppID  = int64(515151)
)

func TestTestControllerHTTPReturnsSeededTestAsJSON(t *testing.T) {
	ctx := context.Background()
	pool := setupHTTPControllerPostgres(t)
	router := setupHTTPControllerRouter(pool)

	testID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	questionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	answerID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	seedHTTPControllerTest(t, ctx, pool, testID, questionID, answerID)

	requestJSON := ``
	expectedJSON := `{
		"id": "11111111-1111-1111-1111-111111111111",
		"name": "HTTP Go Test",
		"description": "Controller integration test",
		"cover": "https://example.com/cover.png",
		"status": "active",
		"starts_at": "2026-05-09T10:00:00Z",
		"ends_at": "2026-05-09T11:00:00Z",
		"ext_link_id": "ext-test-1",
		"attempts": 2,
		"vk_post_id": 987654,
		"questions": [
			{
				"id": "22222222-2222-2222-2222-222222222222",
				"text": "What format do these tests assert?",
				"type": "single_choice",
				"event_id": "11111111-1111-1111-1111-111111111111",
				"order": 1,
				"answers": [
					{
						"id": "33333333-3333-3333-3333-333333333333",
						"question_id": "22222222-2222-2222-2222-222222222222",
						"text": "JSON",
						"is_correct": true,
						"points": 5
					}
				]
			}
		]
	}`

	response := performHTTPControllerRequest(t, router, http.MethodGet, "/api/v1/events/test/"+testID.String(), requestJSON, 7001, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func TestTeamControllerHTTPReturnsSeededTeamsAsJSON(t *testing.T) {
	ctx := context.Background()
	pool := setupHTTPControllerPostgres(t)
	router := setupHTTPControllerRouter(pool)

	eventID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	teamID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	captainID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	memberID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	seedHTTPControllerTeam(t, ctx, pool, eventID, teamID, captainID, memberID)

	requestJSON := ``
	expectedJSON := `[
		{
			"id": "55555555-5555-5555-5555-555555555555",
			"name": "Backend",
			"captain": {
				"id": "66666666-6666-6666-6666-666666666666",
				"vk_id": 7001,
				"firstname": "Ivan",
				"lastname": "Captain",
				"photo": "https://example.com/captain.jpg"
			},
			"members": [
				{
					"id": "77777777-7777-7777-7777-777777777777",
					"vk_id": 7002,
					"firstname": "Petr",
					"lastname": "Member",
					"photo": "https://example.com/member.jpg"
				}
			],
			"total_members": 2
		}
	]`

	response := performHTTPControllerRequest(t, router, http.MethodGet, "/api/v1/events/team/"+eventID.String(), requestJSON, 7001, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func TestBaseEventControllerHTTPGetAllReturnsEmptyJSONList(t *testing.T) {
	ctx := context.Background()
	pool := setupHTTPControllerPostgres(t)
	router := setupHTTPControllerRouter(pool)
	seedHTTPControllerUserWithRole(t, ctx, pool, uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), 7001, "Admin", "User", "https://example.com/admin.jpg", uuid.MustParse("704a0251-f27d-4f54-b09b-e17f8be2d905"))

	requestJSON := ``
	expectedJSON := `[]`

	response := performHTTPControllerRequest(t, router, http.MethodGet, "/api/v1/events", requestJSON, 7001, config.RoleAdmin)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func setupHTTPControllerRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepo := impl.NewUserRepository(pool)
	roleRepo := impl.NewRoleRepositoryPgx(pool)
	vkClient := vkclient.NewVkApiClient("", "https://example.com")
	userService := appservice.NewUserService(userRepo, roleRepo, vkClient)
	eventService := event_service.NewEventService(impl.NewEventBaseRepository(pool), userService, roleRepo)
	eventAsEventService := event_as_event_service.NewEventAsEventService(impl.NewEventEventRepository(pool))
	testService := event_as_test_service.NewEventTestService(impl.NewEventTestRepository(pool))
	teamService := event_as_team_service.NewEventTeamService(impl.NewEventTeamRepository(pool), vkClient)

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	apiV1 := router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(httpControllerSecret, httpControllerAppID))
	ConfigureBaseEventController(apiV1, eventService)
	ConfigureEventController(apiV1, eventService, eventAsEventService, nil)
	ConfigureTestController(apiV1, nil, testService)
	ConfigureTeamEventController(apiV1, nil, teamService, userService)

	return router
}

func performHTTPControllerRequest(t *testing.T, router *gin.Engine, method, target, requestJSON string, vkID int64, role string) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if requestJSON == "" {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader([]byte(requestJSON))
	}

	req := httptest.NewRequest(method, target, body)
	req.Header.Set(controller.HeaderAuthorization, signHTTPControllerToken(vkID, role))
	req.Header.Set(controller.HeaderRequestID, "http-controller-test")
	if requestJSON != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func signHTTPControllerToken(vkID int64, role string) string {
	payload := fmt.Sprintf("%s=%d;%s=%s", config.UserVkIdContextKey, vkID, config.UserRoleContextKey, role)
	ts := int64(1778245200)
	values := url.Values{}
	values.Add("app_id", strconv.FormatInt(httpControllerAppID, 10))
	values.Add("request_id", payload)
	values.Add("ts", strconv.FormatInt(ts, 10))
	values.Add("user_id", strconv.FormatInt(vkID, 10))

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ordered := url.Values{}
	for _, key := range keys {
		ordered.Add(key, values.Get(key))
	}

	mac := hmac.New(sha256.New, []byte(httpControllerSecret))
	mac.Write([]byte(ordered.Encode()))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%d.%s", payload, ts, signature)
}

func setupHTTPControllerPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	container, err := tcPostgres.Run(ctx,
		"postgres:17",
		tcPostgres.WithDatabase("http_controller_test"),
		tcPostgres.WithUsername("http_controller_test"),
		tcPostgres.WithPassword("http_controller_test"),
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
	require.NoError(t, goose.Up(sqlDB, findHTTPControllerMigrationsPath(t)))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))
	t.Cleanup(pool.Close)

	return pool
}

func seedHTTPControllerTest(t *testing.T, ctx context.Context, pool *pgxpool.Pool, testID, questionID, answerID uuid.UUID) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO event_as_tests (id, name, description, cover, status, starts_at, ends_at, ext_link_id, attempts, vk_post_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, testID, "HTTP Go Test", "Controller integration test", "https://example.com/cover.png", "active",
		time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 9, 11, 0, 0, 0, time.UTC),
		"ext-test-1", 2, int64(987654))
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO questions (id, text, type, event_id, q_order)
		VALUES ($1, $2, $3, $4, $5)
	`, questionID, "What format do these tests assert?", "single_choice", testID, 1)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO answers (id, question_id, is_correct, text, points)
		VALUES ($1, $2, $3, $4, $5)
	`, answerID, questionID, true, "JSON", 5)
	require.NoError(t, err)
}

func seedHTTPControllerTeam(t *testing.T, ctx context.Context, pool *pgxpool.Pool, eventID, teamID, captainID, memberID uuid.UUID) {
	t.Helper()

	seedHTTPControllerUser(t, ctx, pool, captainID, 7001, "Ivan", "Captain", "https://example.com/captain.jpg")
	seedHTTPControllerUser(t, ctx, pool, memberID, 7002, "Petr", "Member", "https://example.com/member.jpg")

	_, err := pool.Exec(ctx, `
		INSERT INTO event_as_teams (id, name, status, starts_at, teams_constraint, teams_cap_min, teams_cap_max)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, eventID, "Team Event", "active", time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC), 3, 1, 5)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO teams (id, captain_id, event_id, team_name)
		VALUES ($1, $2, $3, $4)
	`, teamID, captainID, eventID, "Backend")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`, teamID, memberID)
	require.NoError(t, err)
}

func seedHTTPControllerUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, vkID int64, firstName, lastName, photoURL string) {
	t.Helper()

	seedHTTPControllerUserWithRole(t, ctx, pool, id, vkID, firstName, lastName, photoURL, uuid.MustParse("9ef01b95-6d87-4115-80df-7085a647bf36"))
}

func seedHTTPControllerUserWithRole(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, vkID int64, firstName, lastName, photoURL string, roleID uuid.UUID) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, vk_id, first_name, last_name, role_id, photo_url, geo_available, notification_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, vkID, firstName, lastName, roleID, photoURL, false, true)
	require.NoError(t, err)
}

func findHTTPControllerMigrationsPath(t *testing.T) string {
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
