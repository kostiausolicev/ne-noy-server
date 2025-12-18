package repository

import (
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	GetById(id uuid.UUID) (*model.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func (r roleRepository) GetById(id uuid.UUID) (*model.Role, error) {
	var role model.Role
	result := r.db.
		Table("role").
		Select("name").
		Where("id = ?", id).
		First(&role)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return roleRepository{db: db}
}
