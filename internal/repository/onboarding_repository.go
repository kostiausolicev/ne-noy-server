package repository

import (
	"context"
	"ne_noy/internal/model"

	"gorm.io/gorm"
)

type OnboardingRepository interface {
	GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]model.OnboardingModel, error)
	GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]model.OnboardingModel, error)
	SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error
}

type onboardingRepository struct {
	db *gorm.DB
}

func (o *onboardingRepository) GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]model.OnboardingModel, error) {
	var onboardings []model.OnboardingModel
	err := o.db.WithContext(ctx).
		Select("id").
		Where("platform = ?", platform).
		Find(&onboardings).Error
	if err != nil {
		return nil, err
	}
	return onboardings, nil
}

func (o *onboardingRepository) GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]model.OnboardingModel, error) {
	var onboardings []model.OnboardingModel
	err := o.db.WithContext(ctx).
		Table("onboardings o").
		Select("o.*").
		Where(`NOT EXISTS (
			SELECT 1 FROM user_watches_onboardings uwo 
			LEFT JOIN users u ON uwo.user_id = u.id
			WHERE uwo.onboarding_id = o.id AND u.vk_id = ?)
			AND o.platform = ?`, userVkId, platform).
		Find(&onboardings).Error

	if err != nil {
		return nil, err
	}
	return onboardings, nil
}

func (o *onboardingRepository) SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error {
	var user model.User
	err := o.db.WithContext(ctx).
		Table("users").
		Select("users.id").
		Where("vk_id = ?", userVkId).
		First(&user).Error

	if err != nil {
		return err
	}

	err = o.db.WithContext(ctx).
		Table("user_watches_onboardings").
		Create(map[string]interface{}{
			"user_id":       user.ID,
			"onboarding_id": onboardingID,
		}).Error

	return err
}

func NewOnboardingRepository(db *gorm.DB) OnboardingRepository {
	return &onboardingRepository{db: db}
}
