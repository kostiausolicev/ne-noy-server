package model

import "github.com/google/uuid"

type UserWatchesOnboarding struct {
	UserID       uuid.UUID
	User         User
	OnboardingID string
	Onboarding   OnboardingModel
}

func (u UserWatchesOnboarding) TableName() string {
	return "user_watches_onboardings"
}
