package model

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string    `gorm:"size:100;not null"`
	DisplayName string    `gorm:"size:100"`
}
