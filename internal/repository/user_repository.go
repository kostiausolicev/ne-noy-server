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

func (r *userRepository) GetAll() ([]model.User, error) {
	var users []model.User

	result := r.db.Table(`"user"`).
		Select(selectUserFields).
		Scan(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *userRepository) GetAllByFirstNameAndRole(firstName string) ([]model.User, error) {
	var users []model.User
	namePattern := "%" + firstName + "%"

	result := r.db.Table(`"user"`).
		Select(selectUserFields).
		Joins(`LEFT JOIN "role" ON role.id = "user".role_id`).
		Where(`"role".name IN (?, ?)`, config.RoleHikePart, config.RoleAdmin).
		Where(
			r.db.Where(`"user".first_name ILIKE ?`, namePattern).
				Or(`"user".last_name ILIKE ?`, namePattern),
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

	result := r.db.Table(`"user"`).
		Select(selectUserFields).
		Joins(`LEFT JOIN "role" ON role.id = "user".role_id`).
		Where(`"role".name IN (?, ?)`, config.RoleHikePart, config.RoleAdmin).
		Where(
			r.db.Where(`"user".first_name ILIKE ? AND "user".last_name LIKE ?`, fn, ln).
				Or(`"user".first_name ILIKE ? AND "user".last_name LIKE ?`, ln, fn),
		).
		Scan(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
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
