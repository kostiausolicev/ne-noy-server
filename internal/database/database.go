package database

import (
	"log"
	"ne_noy/internal/config"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func Connect(cfg config.DBConfig) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
			Colorful:                  true,
		},
	)

	// Открываем через GORM
	gormDB, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:         newLogger,
		NamingStrategy: schema.NamingStrategy{SingularTable: false},
	})
	if err != nil {
		return nil, err
	}

	// Получаем *sql.DB для goose
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	// Сам запуск миграций
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return nil, err
	}

	// Возвращаем GORM-объект дальше по приложению
	return gormDB, nil
}
