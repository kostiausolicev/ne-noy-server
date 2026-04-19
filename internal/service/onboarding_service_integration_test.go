package service

import (
	"context"
	"database/sql"
	"testing"

	"ne_noy/internal/config"
	repopgx "ne_noy/internal/repository/pgx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnboardingServiceGetAllOnboardingCodesByPlatform(t *testing.T) {
	// Test case: получение всех онбордингов по платформе с разбором JSON-данных
	// Input: в БД есть два onboardings для platform="ios" с валидным JSON в data
	// Expected: сервис возвращает два OnboardingDto с теми же ID и распарсенными данными слайдов без ошибки
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	insertOnboarding(t, sqlDB, "ios-welcome", "ios", "/welcome", `{"slides":[{"media":{"type":"image","blob":"img-1"},"title":"Welcome","subtitle":"Step one"}]}`)
	insertOnboarding(t, sqlDB, "ios-rules", "ios", "/rules", `{"slides":[{"media":{"type":"image","blob":"img-2"},"title":"Rules","subtitle":"Step two"}]}`)

	svc := NewOnboardingService(repopgx.NewOnboardingRepository(pool))

	result, err := svc.GetAllOnboardingCodesByPlatform(context.Background(), "ios")
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "ios-welcome", result[0].ID)
	assert.Equal(t, "Welcome", result[0].Data.Slides[0].Title)
	assert.Equal(t, "ios-rules", result[1].ID)
	assert.Equal(t, "Rules", result[1].Data.Slides[0].Title)
}

func TestOnboardingServiceSetUserOnboardingAndGetOnboardingsForUser(t *testing.T) {
	// Test case: просмотренный онбординг исключается из выдачи для пользователя
	// Input: в БД есть пользователь с vk_id=404 и два onboardings для platform="android", после чего вызывается SetUserOnboarding для одного из них
	// Expected: запись появляется в user_watches_onboardings, а GetOnboardingsForUser возвращает только непросмотренный onboarding
	sqlDB, pool, cleanup := setupTestConnections(t)
	t.Cleanup(cleanup)

	insertUser(t, sqlDB, 404, "Olga", "Ivanova", config.RoleDefault)
	insertOnboarding(t, sqlDB, "android-start", "android", "/start", `{"slides":[{"media":{"type":"image","blob":"img-3"},"title":"Start","subtitle":"Intro"}]}`)
	insertOnboarding(t, sqlDB, "android-finish", "android", "/finish", `{"slides":[{"media":{"type":"image","blob":"img-4"},"title":"Finish","subtitle":"Done"}]}`)

	svc := NewOnboardingService(repopgx.NewOnboardingRepository(pool))

	initial, err := svc.GetOnboardingsForUser(context.Background(), 404, "android")
	require.NoError(t, err)
	require.Len(t, initial, 2)

	err = svc.SetUserOnboarding(context.Background(), 404, "android-start")
	require.NoError(t, err)

	var watches int
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM user_watches_onboardings WHERE onboarding_id = $1`, "android-start").Scan(&watches)
	require.NoError(t, err)
	assert.Equal(t, 1, watches)

	result, err := svc.GetOnboardingsForUser(context.Background(), 404, "android")
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "android-finish", result[0].ID)
	assert.Equal(t, "Finish", result[0].Data.Slides[0].Title)
}

func insertOnboarding(t *testing.T, db *sql.DB, id, platform, path, data string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO onboardings (id, platform, is_active, path, data)
		VALUES ($1, $2, true, $3, $4::jsonb)
	`, id, platform, path, data)
	require.NoError(t, err)
}
