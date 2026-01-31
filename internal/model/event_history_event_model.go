package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventEventHistory struct {
	id   uuid.UUID      `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	data datatypes.JSON `gorm:"type:jsonb"`
}
