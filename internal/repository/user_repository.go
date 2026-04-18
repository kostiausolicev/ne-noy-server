package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// UserRepository defines methods for working with users
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
