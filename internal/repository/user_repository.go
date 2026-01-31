package repository

import (
	"context"
	"errors"
	"ne_noy/internal/config"
	"ne_noy/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetAll(ctx context.Context) ([]model.User, error)
	GetAllByFirstNameAndRole(ctx context.Context, firstName string) ([]model.User, error)
	GetAllByFirstNameAndLastNameAndRole(ctx context.Context, firstName, lastName string) ([]model.User, error)

	GetByVkId(ctx context.Context, vk int64) (*model.User, error)

	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, vkId int64, field string, value interface{}) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error

	ExistEventOrg(ctx context.Context, userId uuid.UUID) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) withScope(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAllByFirstNameAndLastNameAndRole(ctx context.Context, firstName, lastName string) ([]model.User, error) {
	var users []model.User
	fn := "%" + firstName + "%"
	ln := "%" + lastName + "%"
	result := r.withScope(ctx).Table(`"users"`).
		Select(`          users.id,          users.vk_id,          users.first_name,          users.last_name,          users.photo_url       `).
		Joins(`LEFT JOIN roles ON roles.id = users.role_id`).
		Where(`roles.name IN (?, ?)`, config.RoleHikePart, config.RoleAdmin).
		Where(r.db.Where(`users.first_name ILIKE ? AND users.last_name ILIKE ?`, fn, ln).
			Or(`users.first_name ILIKE ? AND users.last_name ILIKE ?`, ln, fn),
		).
		Scan(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.withScope(ctx).
		Select("id", "vk_id", "first_name", "last_name", "photo_url").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetByVkId(ctx context.Context, vk int64) (*model.User, error) {
	var user model.User
	err := r.withScope(ctx).
		Joins("Role").
		Where("vk_id = ?", vk).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.withScope(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, vkId int64, field string, value interface{}) (bool, error) {
	result := r.withScope(ctx).
		Model(&model.User{}).
		Where("vk_id = ?", vkId).
		Update(field, value)

	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.withScope(ctx).
		Where("id = ?", id).
		Delete(&model.User{}).Error
}

func (r *userRepository) GetAllByFirstNameAndRole(ctx context.Context, firstName string) ([]model.User, error) {
	var users []model.User
	pattern := "%" + firstName + "%"

	err := r.withScope(ctx).
		Select("users.id", "users.vk_id", "users.first_name", "users.last_name", "users.photo_url").
		Joins("LEFT JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", []string{config.RoleHikePart, config.RoleAdmin}).
		Where(
			r.db.Where("users.first_name ILIKE ?", pattern).
				Or("users.last_name ILIKE ?", pattern),
		).
		Find(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) ExistEventOrg(ctx context.Context, userId uuid.UUID) (bool, error) {
	var count int64
	err := r.withScope(ctx).
		Table("event_orgs").
		Where("user_id = ?", userId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
