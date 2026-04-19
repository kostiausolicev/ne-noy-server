package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	repopgx "ne_noy/internal/repository/pgx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventServiceCreateEvent(t *testing.T) {
	// Test case: создание события сохраняет основную запись, организаторов и доступные роли
	// Input: в БД есть пользователь-организатор и роль default, вызов CreateEvent с title, startsAt, status, orgs и availableRoles
	// Expected: сервис возвращает созданное событие, а в таблицах events, event_orgs и event_roles появляются связанные записи
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	orgID := insertUserWithPhoto(t, sqlDB, 501, "Oleg", "Organizer", config.RoleDefault, "https://example.com/org.jpg")
	startsAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	title := "Spring Hike"
	status := "DRAFT"

	svc := newEventServiceForTest(pool)

	result, err := svc.CreateEvent(context.Background(), dto.CreateUpdateEventDto{
		Title:          &title,
		StartsAt:       &startsAt,
		Status:         &status,
		Orgs:           []dto.UserMiniDto{{ID: orgID}},
		AvailableRoles: []string{config.RoleDefault},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "Spring Hike", result.Title)
	assert.WithinDuration(t, startsAt, result.StartsAt.UTC(), time.Second)
	assert.Len(t, result.Orgs, 1)
	assert.Equal(t, orgID, result.Orgs[0].ID)

	var (
		dbTitle   string
		dbStatus  string
		dbStartAt time.Time
	)
	err = sqlDB.QueryRow(`SELECT name, status, starts_at FROM events WHERE id = $1`, result.ID).Scan(&dbTitle, &dbStatus, &dbStartAt)
	require.NoError(t, err)
	assert.Equal(t, "Spring Hike", dbTitle)
	assert.Equal(t, "DRAFT", dbStatus)
	assert.WithinDuration(t, startsAt, dbStartAt.UTC(), time.Second)

	var orgCount int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM event_orgs WHERE event_id = $1 AND user_id = $2`, result.ID, orgID).Scan(&orgCount)
	require.NoError(t, err)
	assert.Equal(t, 1, orgCount)

	var roleCount int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*)
		FROM event_roles er
		JOIN roles r ON r.id = er.role_id
		WHERE er.event_id = $1 AND r.name = $2
	`, result.ID, config.RoleDefault).Scan(&roleCount)
	require.NoError(t, err)
	assert.Equal(t, 1, roleCount)
}

func TestEventServiceGetAllForAdmin(t *testing.T) {
	// Test case: администратор получает полный список событий
	// Input: в БД есть пользователь с ролью admin и два события с разными организаторами
	// Expected: GetAll возвращает оба события без ошибки
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	adminID := insertUserWithPhoto(t, sqlDB, 502, "Admin", "User", config.RoleAdmin, "https://example.com/admin.jpg")
	otherOrgID := insertUserWithPhoto(t, sqlDB, 503, "Second", "Org", config.RoleDefault, "https://example.com/org2.jpg")
	firstEventID := insertEvent(t, sqlDB, "First Event", "ACTIVE", time.Now().UTC().Add(2*time.Hour), time.Now().UTC().Add(4*time.Hour), nil, nil)
	secondEventID := insertEvent(t, sqlDB, "Second Event", "ACTIVE", time.Now().UTC().Add(3*time.Hour), time.Now().UTC().Add(5*time.Hour), nil, nil)
	linkEventOrg(t, sqlDB, firstEventID, adminID)
	linkEventOrg(t, sqlDB, secondEventID, otherOrgID)

	svc := newEventServiceForTest(pool)

	result, err := svc.GetAll(context.Background(), 502)
	require.NoError(t, err)
	require.Len(t, result, 2)

	titles := []string{result[0].Title, result[1].Title}
	assert.ElementsMatch(t, []string{"First Event", "Second Event"}, titles)
}

func TestEventServiceGetEventParticipants(t *testing.T) {
	// Test case: получение участников события возвращает пользователя и статус отметки
	// Input: в БД есть событие и два участника в event_participants, один из них отмечен на мероприятии
	// Expected: GetEventParticipants возвращает две записи с корректными user-полями и значением IsChecked
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	eventID := insertEvent(t, sqlDB, "Participants Event", "ACTIVE", time.Now().UTC().Add(2*time.Hour), time.Now().UTC().Add(3*time.Hour), nil, nil)
	firstUserID := insertUserWithPhoto(t, sqlDB, 504, "Nina", "One", config.RoleDefault, "https://example.com/u1.jpg")
	secondUserID := insertUserWithPhoto(t, sqlDB, 505, "Ivan", "Two", config.RoleDefault, "https://example.com/u2.jpg")
	insertParticipantRecord(t, sqlDB, eventID, firstUserID, "app", true, ptrTime(time.Now().UTC()))
	insertParticipantRecord(t, sqlDB, eventID, secondUserID, "app", false, nil)

	svc := newEventServiceForTest(pool)

	result, err := svc.GetEventParticipants(context.Background(), eventID)
	require.NoError(t, err)
	require.Len(t, result, 2)

	checkedByVk := map[int64]bool{}
	for _, item := range result {
		checkedByVk[item.User.VkId] = item.IsChecked
	}

	assert.Equal(t, map[int64]bool{504: true, 505: false}, checkedByVk)
}

func newEventServiceForTest(pool *pgxpool.Pool) EventService {
	roleRepo := repopgx.NewRoleRepositoryPgx(pool)
	userRepo := repopgx.NewUserRepository(pool)
	userSvc := NewUserService(userRepo, roleRepo, stubVkClient{})
	eventRepo := repopgx.NewEventRepositoryPgx(pool)
	return NewEventService(eventRepo, userSvc, roleRepo)
}

func insertEvent(t *testing.T, db *sql.DB, name, status string, startsAt, endsAt time.Time, lat, lon *float64) uuid.UUID {
	t.Helper()

	eventID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO event_as_events (id, name, status, starts_at, ends_at, lat, lon)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, eventID, name, status, startsAt, endsAt, lat, lon)
	require.NoError(t, err)
	return eventID
}

func linkEventOrg(t *testing.T, db *sql.DB, eventID, userID uuid.UUID) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO event_orgs (event_id, user_id, event_type) VALUES ($1, $2, $3)`, eventID, userID, "event")
	require.NoError(t, err)
}

func linkEventRole(t *testing.T, db *sql.DB, eventID uuid.UUID, roleName string) {
	t.Helper()

	roleID := getRoleID(t, db, roleName)
	_, err := db.Exec(`INSERT INTO event_roles (event_id, role_id, event_type) VALUES ($1, $2, $3)`, eventID, roleID, "event")
	require.NoError(t, err)
}

func insertParticipantRecord(t *testing.T, db *sql.DB, eventID, userID uuid.UUID, prepareType string, checked bool, checkTimestamp *time.Time) uuid.UUID {
	t.Helper()

	participantID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO event_participants (id, event_id, user_id, prepare_type, is_checked, check_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, participantID, eventID, userID, prepareType, checked, checkTimestamp)
	require.NoError(t, err)
	return participantID
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
