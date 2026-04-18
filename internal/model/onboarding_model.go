package model

import (
	"time"

	"gorm.io/datatypes"
)

type OnboardingModel struct {
	ID        string
	Platform  string
	IsActive  bool
	Path      string
	Data      datatypes.JSON
	CreatedAt time.Time
	UpdatedAt time.Time
	Users     []User
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
