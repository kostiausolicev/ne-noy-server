package as_activity

import (
	"ne_noy/internal/model/events"

	"gorm.io/datatypes"
)

type AsActivity struct {
	events.EventProfile
	TrainParams         datatypes.JSON
	AvailableActivities datatypes.JSON

	ActivityRecords []UserActivityRecord
}

func (e AsActivity) TableName() string {
	return "event_as_activities"
}
