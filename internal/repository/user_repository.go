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

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetAllByFirstNameAndRole(firstName string) ([]model.User, error)
	GetAllByFirstNameAndLastNameAndRole(firstName, lastName string) ([]model.User, error)
	GetByVkId(vk int64) (*model.User, error)
	Create(user *model.User) (*model.User, error)
	Update(vkId int64, field string, value interface{})
	Delete(id uuid.UUID) error

	GetRole() (*model.Role, error)
	ExistEventOrg(userId uuid.UUID) (bool, error)
}

func (r *userRepository) GetRole() (*model.Role, error) {
	var role model.Role
	result := r.db.Table("roles").
		Select("roles.id, roles.name, roles.display_name").
		Where("roles.name = ?", config.RoleDefault).
		First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *userRepository) GetAll() ([]model.User, error) {
	var users []model.User

	result := r.db.Table(`"users"`).
		Select(`
			users.id,
			users.vk_id,
			users.first_name,
			users.last_name,
			users.photo_url
		`).
		Scan(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) GetAllByFirstNameAndRole(firstName string) ([]model.User, error) {
	var users []model.User
	namePattern := "%" + firstName + "%"

	result := r.db.Table(`"users"`).
		Select(`
			users.id,
			users.vk_id,
			users.first_name,
			users.last_name,
			users.photo_url
		`).
		Joins(`LEFT JOIN roles ON roles.id = users.role_id`).
		Where(`roles.name IN (?, ?)`, config.RoleHikePart, config.RoleAdmin).
		Where(
			r.db.Where(`users.first_name ILIKE ?`, namePattern).
				Or(`users.last_name ILIKE ?`, namePattern),
		).
		Scan(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) GetAllByFirstNameAndLastNameAndRole(firstName, lastName string) ([]model.User, error) {
	var users []model.User
	fn := "%" + firstName + "%"
	ln := "%" + lastName + "%"

	result := r.db.Table(`"users"`).
		Select(`
			users.id,
			users.vk_id,
			users.first_name,
			users.last_name,
			users.photo_url
		`).
		Joins(`LEFT JOIN roles ON roles.id = users.role_id`).
		Where(`roles.name IN (?, ?)`, config.RoleHikePart, config.RoleAdmin).
		Where(
			r.db.Where(`users.first_name ILIKE ? AND users.last_name ILIKE ?`, fn, ln).
				Or(`users.first_name ILIKE ? AND users.last_name ILIKE ?`, ln, fn),
		).
		Scan(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) ExistEventOrg(userId uuid.UUID) (bool, error) {
	result := r.db.Raw(`EXISTS (SELECT 1 FROM event_orgs WHERE user_id = ?) AS has_event`, userId)
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
		Table(`users`).
		Joins("Role").
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
