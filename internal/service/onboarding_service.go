package service

import (
	"context"
	"encoding/json"
	"ne_noy/internal/dto"
	"ne_noy/internal/repository"
)

type OnboardingService interface {
	GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]dto.OnboardingDto, error)
	GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]dto.OnboardingDto, error)
	SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error
}

type onboardingService struct {
	r repository.OnboardingRepository
}

func (o *onboardingService) SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error {
	return o.r.SetUserOnboarding(ctx, userVkId, onboardingID)
}

func (o *onboardingService) GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]dto.OnboardingDto, error) {
	onboardings, err := o.r.GetAllOnboardingCodesByPlatform(ctx, platform)
	if err != nil {
		return nil, err
	}
	onboardingsDto := make([]dto.OnboardingDto, len(onboardings))
	for i, onboarding := range onboardings {
		var data dto.OnboardingData
		err = json.Unmarshal(onboarding.Data, &data)
		if err != nil {
			return nil, err
		}
		onboardingsDto[i] = dto.OnboardingDto{
			ID:   onboarding.ID,
			Data: &data,
		}
	}
	return onboardingsDto, nil
}

func (o *onboardingService) GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]dto.OnboardingDto, error) {
	onboardings, err := o.r.GetOnboardingsForUser(ctx, userVkId, platform)
	if err != nil {
		return nil, err
	}
	onboardingsDto := make([]dto.OnboardingDto, len(onboardings))
	for i, onboarding := range onboardings {
		var data dto.OnboardingData
		err = json.Unmarshal(onboarding.Data, &data)
		if err != nil {
			return nil, err
		}
		onboardingsDto[i] = dto.OnboardingDto{
			ID:       onboarding.ID,
			Platform: onboarding.Platform,
			Path:     onboarding.Path,
			Data:     &data,
		}
	}
	return onboardingsDto, nil
}

func NewOnboardingService(r repository.OnboardingRepository) OnboardingService {
	return &onboardingService{r: r}
}
