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
	result := r.db.Table("role").
		First(&model.Role{}).
		Where("role.name = ?", config.RoleDefault)
	if result.Error != nil {
		return nil, result.Error
	}
	return &model.Role{}, nil
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

type UserRepository interface {
	GetById(id uuid.UUID) (*model.User, error)
	GetByVkId(vk int64) (*model.User, error)
	Create(user *model.User) (*model.User, error)
	Update(user *model.User) (*model.User, error)
	Delete(id uuid.UUID) error

	GetRole() (*model.Role, error)
	ExistEventOrg(userId uuid.UUID) (bool, error)
}

func (r *userRepository) ExistEventOrg(userId uuid.UUID) (bool, error) {
	result := r.db.Raw(`EXISTS (event_org WHERE event_org.user_id = ?) AS hasEvent`, userId)
	if result.Error != nil {
		return false, result.Error
	}
	var hasEvent bool
	result.Scan(&hasEvent)
	return hasEvent, nil
}

func (r *userRepository) GetByVkId(vk int64) (*model.User, error) {
	user := &model.User{}
	result := r.db.
		Preload("Role").
		First(user, "vk_id = ?", vk)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (r *userRepository) GetById(id uuid.UUID) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *userRepository) Delete(id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r *userRepository) Update(user *model.User) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *userRepository) Create(user *model.User) (*model.User, error) {
	//TODO implement me
	panic("implement me")
}
