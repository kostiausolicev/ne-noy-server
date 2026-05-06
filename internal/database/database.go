package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"ne_noy/internal/config"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

func Connect(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	dsn := cfg.DSN()

	// 1. Сначала запускаем миграции через database/sql
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// 2. Настраиваем pgxpool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	// Настройки пула
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	// 3. Логирование SQL-запросов
	poolConfig.ConnConfig.Tracer = &QueryTracer{}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql db for goose: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

type QueryTracer struct {
	logger *log.Logger
}

func (t *QueryTracer) ensureLogger() {
	if t.logger == nil {
		t.logger = log.New(os.Stdout, "[PGX] ", log.LstdFlags)
	}
}

func (t *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.ensureLogger()

	// TODO добавить формирование строки
	t.logger.Printf("START SQL: %s | args=%v", data.SQL, data.Args)
	return context.WithValue(ctx, "query_start_time", time.Now())
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	t.ensureLogger()

	start, ok := ctx.Value("query_start_time").(time.Time)
	if ok {
		duration := time.Since(start)
		if data.Err != nil {
			t.logger.Printf("ERROR SQL (%s): %v", duration, data.Err)
		} else {
			t.logger.Printf("END SQL (%s)", duration)
		}
	} else {
		if data.Err != nil {
			t.logger.Printf("ERROR SQL: %v", data.Err)
		} else {
			t.logger.Printf("END SQL")
		}
	}
}
