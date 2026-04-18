package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventEventHistory struct {
	id   uuid.UUID
	data datatypes.JSON
}
