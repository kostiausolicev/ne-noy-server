package repository

import (
	"context"
	"ne_noy/internal/model"
)

// OnboardingRepository describes persistence operations for onboarding content.
type OnboardingRepository interface {
	// GetAllOnboardingCodesByPlatform returns all onboarding entries for a platform.
	GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]model.OnboardingModel, error)

	// GetOnboardingsForUser returns onboarding entries available to a user on a platform.
	GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]model.OnboardingModel, error)

	// SetUserOnboarding marks an onboarding item as viewed by a user.
	SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error
}
