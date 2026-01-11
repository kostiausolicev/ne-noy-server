package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type EventQueueModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	PostID      int64          `gorm:"type:bigint;not null"`
	Text        string         `gorm:"type:text"`
	Lat         *float64       `gorm:"type:decimal(10,8)"`
	Lon         *float64       `gorm:"type:decimal(11,8)"`
	Address     *string        `gorm:"type:text"`
	Poll        datatypes.JSON `gorm:"type:jsonb"`
	Photos      datatypes.JSON `gorm:"type:jsonb"`
	Attachments datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
}

func (e EventQueueModel) TableName() string {
	return "queue_events"
}
