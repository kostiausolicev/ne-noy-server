package repository

import (
	"context"
	"ne_noy/internal/config"
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func (r *roleRepository) withScope(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.withScope(ctx).
		Select("id", "name", "display_name").
		Where("name = ?", config.RoleDefault).
		First(&role).Error

	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetById(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	result := r.withScope(ctx).
		Table("roles").
		Select("name").
		Where("id = ?", id).
		First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}
