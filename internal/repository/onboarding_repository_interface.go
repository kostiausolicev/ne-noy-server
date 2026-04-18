package repository

import (
	"context"
	"ne_noy/internal/model"
)

// OnboardingRepository
type OnboardingRepository interface {
	GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]model.OnboardingModel, error)
	GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]model.OnboardingModel, error)
	SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error
}
