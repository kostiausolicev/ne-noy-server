package as_activity

import (
	"ne_noy/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type UserActivityRecord struct {
	model.BaseModel
	UserID        *uuid.UUID
	User          *model.User
	ActivityID    *uuid.UUID
	Activity      *string
	ActivityEvent *AsActivity `gorm:"foreignKey:ActivityID"`
	Starts        *time.Time
	Ends          *time.Time
	ParamValues   datatypes.JSON
	Hash          *string
}

func (u UserActivityRecord) TableName() string {
	return "user_activity_records"
}
