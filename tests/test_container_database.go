package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm/schema"
)

func findMigrationsPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cur := wd
	for i := 0; i < 6; i++ {
		p := filepath.Join(cur, "migrations")
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p, nil
		}
		cur = filepath.Join(cur, "..")
	}
	return "", fmt.Errorf("migrations directory not found from %s", wd)
}

func SetupPostgres(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Запускаем контейнер
	container, err := tcPostgres.Run(ctx,
		"postgres:17", // или postgres:16 / postgres:17-alpine — 17 иногда медленнее
		tcPostgres.WithDatabase("test"),
		tcPostgres.WithUsername("test"),
		tcPostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2). // иногда нужно 2 раза (primary + standby в некоторых образах)
				WithStartupTimeout(60*time.Second),
		),
		// Дополнительно — ждём, пока порт реально открыт и отвечает
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(45*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate postgres container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: false},
	})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("failed to get sql DB from gorm: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	migrationsPath, err := findMigrationsPath()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Сам запуск миграций
	if err := goose.Up(sqlDB, migrationsPath); err != nil {
		wd, _ := os.Getwd()
		fmt.Printf("working dir: %s\n", wd)
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Возвращаем GORM-объект дальше по приложению
	return gormDB
}
