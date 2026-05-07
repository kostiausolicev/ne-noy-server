package impl

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_test"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEventTestRepositoryCreateQuestionAnswerAndUserAnswer(t *testing.T) {
	ctx := context.Background()
	pool := setupTestRepositoryPostgres(t)
	repo := NewEventTestRepository(pool)

	userID := seedTestRepoUser(t, ctx, pool, 3001, "Ivan", "Tester")
	test, err := repo.CreateTest(ctx, &as_test.AsTest{
		EventProfile: events.EventProfile{
			Name:     "Go Basics",
			Status:   "active",
			StartsAt: time.Now().UTC(),
		},
		Attempts: 2,
	})
	require.NoError(t, err)
	require.Equal(t, "Go Basics", test.Name)
	require.Empty(t, test.Questions)

	question, err := repo.AddQuestion(ctx, test.ID, as_test.Question{
		Text:   "What keyword starts a goroutine?",
		Type:   "single_choice",
		QOrder: 1,
	})
	require.NoError(t, err)
	require.Equal(t, test.ID, question.EventID)

	answer, err := repo.AddAnswer(ctx, question.ID, as_test.Answer{
		Text:      "go",
		IsCorrect: true,
		Points:    5,
	})
	require.NoError(t, err)
	require.Equal(t, question.ID, answer.QuestionID)

	savedAnswer, err := repo.SetUserAnswer(ctx, as_test.UserAnswer{
		UserID:     userID,
		QuestionID: question.ID,
		AnswerID:   &answer.ID,
	})
	require.NoError(t, err)
	require.Equal(t, 5, savedAnswer.Points)

	fullTest, err := repo.GetTest(ctx, test.ID)
	require.NoError(t, err)
	require.Len(t, fullTest.Questions, 1)
	require.Len(t, fullTest.Questions[0].Answers, 1)
	require.Equal(t, "go", fullTest.Questions[0].Answers[0].Text)
}

func TestEventTestRepositoryUpdateTest(t *testing.T) {
	ctx := context.Background()
	pool := setupTestRepositoryPostgres(t)
	repo := NewEventTestRepository(pool)

	test, err := repo.CreateTest(ctx, &as_test.AsTest{
		EventProfile: events.EventProfile{
			Name:     "Old name",
			Status:   "draft",
			StartsAt: time.Now().UTC(),
		},
		Attempts: 1,
	})
	require.NoError(t, err)

	description := "Updated description"
	updated, err := repo.UpdateTest(ctx, test.ID, as_test.AsTest{
		EventProfile: events.EventProfile{
			Name:        "New name",
			Description: &description,
			Status:      "active",
		},
		Attempts: 3,
	})
	require.NoError(t, err)
	require.Equal(t, "New name", updated.Name)
	require.Equal(t, "active", updated.Status)
	require.Equal(t, 3, updated.Attempts)
	require.NotNil(t, updated.Description)
	require.Equal(t, description, *updated.Description)
}

func setupTestRepositoryPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	container, err := tcPostgres.Run(ctx,
		"postgres:17",
		tcPostgres.WithDatabase("test_repo_test"),
		tcPostgres.WithUsername("test_repo_test"),
		tcPostgres.WithPassword("test_repo_test"),
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
	require.NoError(t, goose.Up(sqlDB, findTestRepoMigrationsPath(t)))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))
	t.Cleanup(pool.Close)

	return pool
}

func seedTestRepoUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, vkID int64, firstName, lastName string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, vk_id, first_name, last_name, photo_url)
		VALUES ($1, $2, $3, $4, $5)
	`, id, vkID, firstName, lastName, "")
	require.NoError(t, err)

	return id
}

func findTestRepoMigrationsPath(t *testing.T) string {
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
