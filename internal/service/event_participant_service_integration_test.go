package service

import (
	"context"
	"testing"
	"time"

	"ne_noy/internal/apperror"
	"ne_noy/internal/config"
	"ne_noy/internal/dto"
	repopgx "ne_noy/internal/repository/pgx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventParticipantServiceParticipantToEvent(t *testing.T) {
	// Test case: пользователь с разрешённой ролью записывается на событие
	// Input: в БД есть пользователь с ролью default, событие и запись event_roles с той же ролью, вызов ParticipantToEvent
	// Expected: сервис возвращает true, а в event_participants появляется запись для пользователя и события
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	userID := insertUserWithPhoto(t, sqlDB, 601, "Pavel", "Participant", config.RoleDefault, "https://example.com/p1.jpg")
	eventID := insertEvent(t, sqlDB, "Join Event", "ACTIVE", time.Now().UTC().Add(time.Hour), time.Now().UTC().Add(2*time.Hour), nil, nil)
	linkEventRole(t, sqlDB, eventID, config.RoleDefault)

	svc := newEventParticipantServiceForTest(pool)

	ok, err := svc.ParticipantToEvent(context.Background(), eventID, 601, "app")
	require.NoError(t, err)
	assert.True(t, ok)

	var count int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE event_id = $1 AND user_id = $2`, eventID, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestEventParticipantServiceUnParticipantToEvent(t *testing.T) {
	// Test case: пользователь снимает участие в событии
	// Input: в БД уже есть запись event_participants для пользователя с vk_id=602 и события, вызов UnParticipantToEvent
	// Expected: сервис возвращает true, а запись участника удаляется из БД
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	userID := insertUserWithPhoto(t, sqlDB, 602, "Mila", "Leave", config.RoleDefault, "https://example.com/p2.jpg")
	eventID := insertEvent(t, sqlDB, "Leave Event", "ACTIVE", time.Now().UTC().Add(time.Hour), time.Now().UTC().Add(2*time.Hour), nil, nil)
	insertParticipantRecord(t, sqlDB, eventID, userID, "app", false, nil)

	svc := newEventParticipantServiceForTest(pool)

	ok, err := svc.UnParticipantToEvent(context.Background(), eventID, 602)
	require.NoError(t, err)
	assert.True(t, ok)

	var count int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE event_id = $1 AND user_id = $2`, eventID, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestEventParticipantServiceCheckParticipantByEventQR(t *testing.T) {
	// Test case: отметка по QR добавляет отсутствующего участника и помечает его как checked
	// Input: в БД есть событие с координатами, пользователь с ролью default и доступной ролью события, вызов CheckParticipant с типом "Event QR" и координатами рядом с событием
	// Expected: сервис возвращает nil, а в event_participants создаётся или обновляется запись с is_checked=true, check_type="Event QR" и ненулевым временем отметки
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	lat := 55.7558
	lon := 37.6173
	userID := insertUserWithPhoto(t, sqlDB, 603, "Kirill", "QR", config.RoleDefault, "https://example.com/p3.jpg")
	eventID := insertEvent(t, sqlDB, "QR Event", "ACTIVE", time.Now().UTC().Add(time.Hour), time.Now().UTC().Add(2*time.Hour), &lat, &lon)
	linkEventRole(t, sqlDB, eventID, config.RoleDefault)

	checkTime := time.Now().UTC().Truncate(time.Second)
	svc := newEventParticipantServiceForTest(pool)

	err := svc.CheckParticipant(context.Background(), dto.CheckEventParticipant{
		UserId:    userID,
		EventId:   eventID,
		Timestamp: checkTime,
		CheckType: "Event QR",
		Lat:       floatPtr(55.75581),
		Long:      floatPtr(37.61731),
	})
	require.NoError(t, err)

	var (
		isChecked     bool
		checkType     string
		storedPrepare string
		storedTime    time.Time
	)
	err = sqlDB.QueryRow(`
		SELECT is_checked, check_type, prepare_type, check_timestamp
		FROM event_participants
		WHERE event_id = $1 AND user_id = $2
	`, eventID, userID).Scan(&isChecked, &checkType, &storedPrepare, &storedTime)
	require.NoError(t, err)
	assert.True(t, isChecked)
	assert.Equal(t, "Event QR", checkType)
	assert.Equal(t, "app", storedPrepare)
	assert.Equal(t, checkTime, storedTime.UTC())
}

func TestEventParticipantServiceCheckParticipantByAdmin(t *testing.T) {
	// Test case: организатор отмечает существующего участника через админскую панель
	// Input: в БД есть событие, организатор в event_orgs, участник в event_participants и вызов CheckParticipant с типом "Admin panel" и vk_id автора отметки
	// Expected: сервис возвращает nil, а в записи участника проставляются is_checked=true, check_author и тип отметки
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	orgID := insertUserWithPhoto(t, sqlDB, 604, "Olga", "Org", config.RoleDefault, "https://example.com/org-admin.jpg")
	userID := insertUserWithPhoto(t, sqlDB, 605, "Sergey", "Member", config.RoleDefault, "https://example.com/member.jpg")
	eventID := insertEvent(t, sqlDB, "Admin Check Event", "ACTIVE", time.Now().UTC().Add(time.Hour), time.Now().UTC().Add(2*time.Hour), nil, nil)
	linkEventOrg(t, sqlDB, eventID, orgID)
	insertParticipantRecord(t, sqlDB, eventID, userID, "app", false, nil)

	checkAuthorVkID := int64(604)
	checkTime := time.Now().UTC().Truncate(time.Second)
	svc := newEventParticipantServiceForTest(pool)

	err := svc.CheckParticipant(context.Background(), dto.CheckEventParticipant{
		UserId:          userID,
		EventId:         eventID,
		CheckAuthorVkId: &checkAuthorVkID,
		Timestamp:       checkTime,
		CheckType:       "Admin panel",
	})
	require.NoError(t, err)

	var (
		isChecked   bool
		checkType   string
		checkAuthor uuid.UUID
	)
	err = sqlDB.QueryRow(`
		SELECT is_checked, check_type, check_author
		FROM event_participants
		WHERE event_id = $1 AND user_id = $2
	`, eventID, userID).Scan(&isChecked, &checkType, &checkAuthor)
	require.NoError(t, err)
	assert.True(t, isChecked)
	assert.Equal(t, "Admin panel", checkType)
	assert.Equal(t, orgID, checkAuthor)
}

func TestEventParticipantServiceCheckParticipantByEventQRTooFar(t *testing.T) {
	// Test case: отметка по QR отклоняется, если пользователь слишком далеко от координат события
	// Input: в БД есть событие с координатами, пользователь с разрешённой ролью и вызов CheckParticipant с типом "Event QR" и удалённой геопозицией
	// Expected: сервис возвращает ошибку ParticipantLocationTooLageErr и не создаёт запись в event_participants
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	lat := 55.7558
	lon := 37.6173
	userID := insertUserWithPhoto(t, sqlDB, 606, "Far", "Away", config.RoleDefault, "https://example.com/p4.jpg")
	eventID := insertEvent(t, sqlDB, "Far Event", "ACTIVE", time.Now().UTC().Add(time.Hour), time.Now().UTC().Add(2*time.Hour), &lat, &lon)
	linkEventRole(t, sqlDB, eventID, config.RoleDefault)

	svc := newEventParticipantServiceForTest(pool)

	err := svc.CheckParticipant(context.Background(), dto.CheckEventParticipant{
		UserId:    userID,
		EventId:   eventID,
		Timestamp: time.Now().UTC(),
		CheckType: "Event QR",
		Lat:       floatPtr(59.9343),
		Long:      floatPtr(30.3351),
	})
	require.ErrorIs(t, err, apperror.ParticipantLocationTooLageErr)

	var count int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE event_id = $1 AND user_id = $2`, eventID, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func newEventParticipantServiceForTest(pool *pgxpool.Pool) EventParticipantService {
	eventRepo := repopgx.NewEventRepositoryPgx(pool)
	participantRepo := repopgx.NewEventParticipantRepository(pool)
	return NewEventParticipantService(participantRepo, eventRepo, 300)
}

func floatPtr(value float64) *float64 {
	return &value
}
