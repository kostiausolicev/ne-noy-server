package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventQueueModel struct {
	ID          uuid.UUID
	PostID      int64
	Text        string
	Lat         *float64
	Lon         *float64
	Address     *string
	Poll        datatypes.JSON
	Photos      datatypes.JSON
	Attachments datatypes.JSON
	CreatedAt   time.Time
}

func (e EventQueueModel) TableName() string {
	return "queue_events"
}
