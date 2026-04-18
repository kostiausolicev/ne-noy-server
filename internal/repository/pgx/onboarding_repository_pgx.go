package pgx

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type onboardingRepository struct {
	pool *pgxpool.Pool
}

func NewOnboardingRepository(pool *pgxpool.Pool) repository.OnboardingRepository {
	return &onboardingRepository{pool: pool}
}

func (o *onboardingRepository) GetAllOnboardingCodesByPlatform(ctx context.Context, platform string) ([]model.OnboardingModel, error) {
	rows, err := o.pool.Query(ctx, `SELECT id FROM onboardings WHERE platform = $1`, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.OnboardingModel, 0)
	for rows.Next() {
		var m model.OnboardingModel
		if err := rows.Scan(&m.ID); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}

func (o *onboardingRepository) GetOnboardingsForUser(ctx context.Context, userVkId int64, platform string) ([]model.OnboardingModel, error) {
	rows, err := o.pool.Query(ctx, `
		SELECT o.id, o.platform, o.is_active, o.path, o.data, o.created_at, o.updated_at FROM onboardings o
		WHERE o.platform = $2
		AND NOT EXISTS (
			SELECT 1 FROM user_watches_onboardings uwo
			LEFT JOIN users u ON uwo.user_id = u.id
			WHERE uwo.onboarding_id = o.id AND u.vk_id = $1
		)
	`, userVkId, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]model.OnboardingModel, 0)
	for rows.Next() {
		var m model.OnboardingModel
		// Scan fields: id, platform, is_active, path, data, created_at, updated_at
		if err := rows.Scan(&m.ID, &m.Platform, &m.IsActive, &m.Path, &m.Data, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil
}

func (o *onboardingRepository) SetUserOnboarding(ctx context.Context, userVkId int64, onboardingID string) error {
	var userId uuid.UUID
	row := o.pool.QueryRow(ctx, `SELECT id FROM users WHERE vk_id = $1`, userVkId)
	if err := row.Scan(&userId); err != nil {
		return err
	}

	_, err := o.pool.Exec(ctx, `INSERT INTO user_watches_onboardings (user_id, onboarding_id) VALUES ($1, $2)`, userId, onboardingID)
	return err
}
