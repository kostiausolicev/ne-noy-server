package database

import (
	"log"
	"ne_noy/internal/config"
	"ne_noy/internal/model"
	"os"
	"time"

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
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:         newLogger,
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.EventParticipant{},
		&model.Event{},
		&model.EventAttachment{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}
