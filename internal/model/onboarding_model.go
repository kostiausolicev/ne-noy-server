package model

import (
	"time"

	"gorm.io/datatypes"
)

type OnboardingModel struct {
	ID        string         `gorm:"type:varchar(50);column:id;primaryKey"`
	Platform  string         `gorm:"type:varchar(50);column:platform"`
	IsActive  bool           `gorm:"type:boolean;column:is_active"`
	Path      string         `gorm:"type:varchar(255);column:path"`
	Data      datatypes.JSON `gorm:"type:jsonb;column:data"`
	CreatedAt time.Time      `gorm:"type:timestamp;column:created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamp;column:updated_at"`
	Users     []User         `gorm:"many2many:user_watches_onboardings;"`
}

func (o OnboardingModel) TableName() string {
	return "onboardings"
}

type OnboardingData struct {
	Slides []OnboardingSlide `json:"slides"`
}

type OnboardingSlide struct {
	Media    SlideMedia `json:"media"`
	Title    string     `json:"title"`
	Subtitle string     `json:"subtitle"`
}

type SlideMedia struct {
	Type string `json:"type"`
	Blob string `json:"blob"`
}
