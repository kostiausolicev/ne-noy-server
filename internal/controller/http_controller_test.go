package controller_test

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
	controller "ne_noy/internal/controller"
	"ne_noy/internal/controller/middleware"
	"ne_noy/internal/repository/impl"
	appservice "ne_noy/internal/service"
	event_as_event_service "ne_noy/internal/service/event/event_as_event"

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
	controllerHTTPSecret = "controller-http-secret"
	controllerHTTPAppID  = int64(616161)
)

func TestServiceControllerHTTPHealthReturnsJSON(t *testing.T) {
	pool := setupControllerHTTPPostgres(t)
	router := setupControllerHTTPRouter(pool)

	requestJSON := ``
	expectedJSON := `{"status": "ok"}`

	response := performControllerHTTPRequest(t, router, http.MethodGet, "/health", requestJSON, 8001, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func TestServiceControllerHTTPValidateAcceptsSignedRequest(t *testing.T) {
	pool := setupControllerHTTPPostgres(t)
	router := setupControllerHTTPRouter(pool)

	requestJSON := ``
	expectedBody := ``

	response := performControllerHTTPRequest(t, router, http.MethodGet, "/api/v1/validate", requestJSON, 8001, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, expectedBody, response.Body.String())
}

func TestUserControllerHTTPGetByVkIDReturnsSeededUserAsJSON(t *testing.T) {
	ctx := context.Background()
	pool := setupControllerHTTPPostgres(t)
	router := setupControllerHTTPRouter(pool)

	userID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	seedControllerHTTPUser(t, ctx, pool, userID, 8001, "Olga", "User", "https://example.com/user.jpg")

	requestJSON := ``
	expectedJSON := `{
		"id": "88888888-8888-8888-8888-888888888888",
		"vkId": 8001,
		"firstName": "Olga",
		"lastName": "User",
		"photo": "https://example.com/user.jpg",
		"role": {
			"id": "9ef01b95-6d87-4115-80df-7085a647bf36",
			"name": "default",
			"displayName": "Внешний участник"
		},
		"geoAvailable": false,
		"notificationAvailable": true
	}`

	response := performControllerHTTPRequest(t, router, http.MethodGet, "/api/v1/users/vk/8001", requestJSON, 8001, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func TestOnboardingControllerHTTPGetForUserReturnsUnwatchedOnboardingsAsJSON(t *testing.T) {
	ctx := context.Background()
	pool := setupControllerHTTPPostgres(t)
	router := setupControllerHTTPRouter(pool)

	userID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	seedControllerHTTPUser(t, ctx, pool, userID, 8002, "Anna", "Learner", "https://example.com/learner.jpg")
	seedControllerHTTPOnboarding(t, ctx, pool)

	requestJSON := ``
	expectedJSON := `[
		{
			"id": "intro",
			"platform": "vk",
			"path": "/welcome",
			"data": {
				"slides": [
					{
						"media": {
							"type": "image",
							"blob": "welcome.png"
						},
						"title": "Welcome",
						"subtitle": "Start here"
					}
				]
			}
		}
	]`

	response := performControllerHTTPRequest(t, router, http.MethodGet, "/api/v1/onboardings?platform=vk", requestJSON, 8002, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func TestEventQueueControllerHTTPGetAllReturnsSeededQueueAsJSON(t *testing.T) {
	ctx := context.Background()
	pool := setupControllerHTTPPostgres(t)
	router := setupControllerHTTPRouter(pool)

	seedControllerHTTPQueueEvent(t, ctx, pool)

	requestJSON := ``
	expectedJSON := `[
		{
			"post_id": 9001,
			"text": "Queue post",
			"poll": {
				"id": 42,
				"answers": [
					{
						"id": 1,
						"text": "Yes"
					}
				]
			},
			"photos": [
				{
					"id": 77,
					"sizes": [
						{
							"type": "x",
							"url": "https://example.com/photo.jpg"
						}
					]
				}
			],
			"attachments": [
				{
					"id": 88,
					"url": "https://example.com/doc.pdf",
					"title": "Rules"
				}
			]
		}
	]`

	response := performControllerHTTPRequest(t, router, http.MethodGet, "/api/v1/events/queue", requestJSON, 8001, config.RoleAdmin)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, expectedJSON, response.Body.String())
}

func TestVkCallbackControllerHTTPConfirmationReturnsExpectedBody(t *testing.T) {
	pool := setupControllerHTTPPostgres(t)
	router := setupControllerHTTPRouter(pool)

	requestJSON := `{
		"type": "confirmation",
		"group_id": 123,
		"secret": "controller-http-secret",
		"object": {}
	}`
	expectedBody := `0bf0212a`

	response := performControllerHTTPRequest(t, router, http.MethodPost, "/api/v1/callback", requestJSON, 8001, config.RoleDefault)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, expectedBody, response.Body.String())
}

func setupControllerHTTPRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepo := impl.NewUserRepository(pool)
	roleRepo := impl.NewRoleRepositoryPgx(pool)
	eventQueueRepo := impl.NewEventQueueRepository(pool)
	eventEventRepo := impl.NewEventEventRepository(pool)
	vkClient := vkclient.NewVkApiClient("", "https://example.com")

	userService := appservice.NewUserService(userRepo, roleRepo, vkClient)
	onboardingService := appservice.NewOnboardingService(impl.NewOnboardingRepository(pool))
	eventQueueService := appservice.NewEventQueueService(eventQueueRepo)
	eventParticipantService := event_as_event_service.NewEventParticipantService(impl.NewEventParticipantRepository(pool), eventEventRepo, 100)
	vkCallbackService := appservice.NewVkCallbackService(eventQueueRepo, eventEventRepo, eventParticipantService)

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	controller.ConfigureServiceController(router.Group("/"), userRepo)

	apiV1 := router.Group("/api/v1")
	controller.ConfigureVkCallBackController(apiV1, controllerHTTPSecret, vkCallbackService)
	apiV1.Use(middleware.AuthMiddleware(controllerHTTPSecret, controllerHTTPAppID))
	controller.ApiServiceController(apiV1)
	controller.ConfigureUserController(apiV1, userService)
	controller.ConfigureOnboardingController(apiV1, onboardingService)
	apiV1.Use(middleware.AdminMiddleware())
	controller.ConfigureEventQueueController(apiV1, eventQueueService)
	controller.ConfigureAdminUserController(apiV1, userService)

	return router
}

func performControllerHTTPRequest(t *testing.T, router *gin.Engine, method, target, requestJSON string, vkID int64, role string) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if requestJSON == "" {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader([]byte(requestJSON))
	}

	req := httptest.NewRequest(method, target, body)
	req.Header.Set(controller.HeaderAuthorization, signControllerHTTPToken(vkID, role))
	req.Header.Set(controller.HeaderRequestID, "controller-http-test")
	if requestJSON != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func signControllerHTTPToken(vkID int64, role string) string {
	payload := fmt.Sprintf("%s=%d;%s=%s", config.UserVkIdContextKey, vkID, config.UserRoleContextKey, role)
	ts := int64(1778245200)
	values := url.Values{}
	values.Add("app_id", strconv.FormatInt(controllerHTTPAppID, 10))
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

	mac := hmac.New(sha256.New, []byte(controllerHTTPSecret))
	mac.Write([]byte(ordered.Encode()))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s.%d.%s", payload, ts, signature)
}

func setupControllerHTTPPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	container, err := tcPostgres.Run(ctx,
		"postgres:17",
		tcPostgres.WithDatabase("controller_http_test"),
		tcPostgres.WithUsername("controller_http_test"),
		tcPostgres.WithPassword("controller_http_test"),
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
	require.NoError(t, goose.Up(sqlDB, findControllerHTTPMigrationsPath(t)))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))
	t.Cleanup(pool.Close)

	return pool
}

func seedControllerHTTPUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, vkID int64, firstName, lastName, photoURL string) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, vk_id, first_name, last_name, role_id, photo_url, geo_available, notification_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, vkID, firstName, lastName, uuid.MustParse("9ef01b95-6d87-4115-80df-7085a647bf36"), photoURL, false, true)
	require.NoError(t, err)
}

func seedControllerHTTPOnboarding(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO onboardings (id, platform, is_active, path, data)
		VALUES ($1, $2, $3, $4, $5)
	`, "intro", "vk", true, "/welcome", `{"slides":[{"media":{"type":"image","blob":"welcome.png"},"title":"Welcome","subtitle":"Start here"}]}`)
	require.NoError(t, err)
}

func seedControllerHTTPQueueEvent(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO queue_events (id, post_id, text, poll, photos, attachments)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), 9001, "Queue post",
		`[{"id":42,"question":"Join?","answers":[{"id":1,"text":"Yes"}]}]`,
		`[{"id":77,"sizes":[{"type":"x","url":"https://example.com/photo.jpg"}]}]`,
		`[{"id":88,"title":"Rules","url":"https://example.com/doc.pdf"}]`)
	require.NoError(t, err)
}

func findControllerHTTPMigrationsPath(t *testing.T) string {
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
