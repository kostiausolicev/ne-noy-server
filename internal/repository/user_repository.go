package repository

import (
	"errors"
	"ne_noy/internal/config"
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) GetRole() (*model.Role, error) {
	var role model.Role
	result := r.db.Table("role").
		Select("role.id, role.name, role.display_name").
		Where("role.name = ?", config.RoleDefault).
		First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // можно отдельно обработать создание роли
		}
		return nil, result.Error
	}
	return &role, nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

type UserRepository interface {
	GetByVkId(vk int64) (*model.User, error)
	Create(user *model.User) (*model.User, error)
	Update(vkId int64, field string, value interface{})
	Delete(id uuid.UUID) error

	GetRole() (*model.Role, error)
	ExistEventOrg(userId uuid.UUID) (bool, error)
}

func (r *userRepository) ExistEventOrg(userId uuid.UUID) (bool, error) {
	result := r.db.Raw(`EXISTS (event_org WHERE event_org.user_id = ?) AS has_event`, userId)
	if result.Error != nil {
		return false, result.Error
	}
	var hasEvent bool
	result.Scan(&hasEvent)
	return hasEvent, nil
}

func (r *userRepository) GetByVkId(vk int64) (*model.User, error) {
	var user model.User
	result := r.db.
		Preload("Role").
		Where("vk_id = ?", vk).
		First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r *userRepository) Update(vkId int64, field string, value interface{}) {
	r.db.Model(&model.User{}).
		Where("vk_id = ?", vkId).
		Update(field, value)
}

func (r *userRepository) Create(user *model.User) (*model.User, error) {
	r.db.Create(user)
	return user, nil
}
