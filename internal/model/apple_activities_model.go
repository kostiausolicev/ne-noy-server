package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserActivityRecord struct {
	ID              uuid.UUID
	UserID          uuid.UUID `db:"user_id"`
	User            User
	EventActivityID uuid.UUID `db:"activity_id"`
	EventActivity   EventAsActivity
	Activity        string
	Starts          time.Time
	Ends            time.Time
	ParamValues     json.RawMessage // []ActivityParamValues
	CreatedAt       time.Time
	Hash            string
}

type ActivityParamValues struct {
	Param string
	Value any
}
