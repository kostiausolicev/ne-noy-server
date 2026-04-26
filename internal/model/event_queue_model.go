package model

import (
	"gorm.io/datatypes"
)

type EventQueueModel struct {
	BaseModel
	PostID      int64
	Text        string
	Lat         *float64
	Lon         *float64
	Address     *string
	Poll        datatypes.JSON
	Photos      datatypes.JSON
	Attachments datatypes.JSON
}

func (e EventQueueModel) TableName() string {
	return "queue_events"
}
