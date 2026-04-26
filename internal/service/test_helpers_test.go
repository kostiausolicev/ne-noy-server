package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func findMigrationsPath(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	cur := wd
	for i := 0; i < 6; i++ {
		path := filepath.Join(cur, "migrations")
		if info, statErr := os.Stat(path); statErr == nil && info.IsDir() {
			return path
		}
		cur = filepath.Join(cur, "..")
	}

	t.Fatalf("migrations directory not found from %s", wd)
	return ""
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	sqlDB, _, cleanup := setupTestConnections(t)
	return sqlDB, cleanup
}

func truncateAllTables(t *testing.T, db *sql.DB) {
	t.Helper()

	rows, err := db.Query(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		  AND tablename <> 'goose_db_version'
	`)
	require.NoError(t, err)
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		require.NoError(t, rows.Scan(&table))
		tables = append(tables, fmt.Sprintf(`"%s"`, table))
	}
	require.NoError(t, rows.Err())

	if len(tables) == 0 {
		return
	}

	_, err = db.Exec("TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func startTestPostgres(t *testing.T) (string, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := tcPostgres.Run(ctx,
		"postgres:17",
		tcPostgres.WithDatabase("as_test"),
		tcPostgres.WithUsername("as_test"),
		tcPostgres.WithPassword("as_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	return connStr, func() {
		require.NoError(t, container.Terminate(ctx))
	}
}

func setupTestConnections(t *testing.T) (*sql.DB, *pgxpool.Pool, func()) {
	t.Helper()

	connStr, stopContainer := startTestPostgres(t)
	ctx := context.Background()

	sqlDB, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	require.NoError(t, sqlDB.PingContext(ctx))

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(sqlDB, findMigrationsPath(t)))

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		truncateAllTables(t, sqlDB)
		require.NoError(t, sqlDB.Close())
		stopContainer()
	}

	return sqlDB, pool, cleanup
}
