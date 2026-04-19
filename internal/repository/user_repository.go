package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// UserRepository describes persistence operations for users.
type UserRepository interface {
	// GetAll returns all users.
	GetAll(ctx context.Context) ([]model.User, error)

	// GetAllByFirstNameAndRole returns users filtered by first name and role constraints.
	GetAllByFirstNameAndRole(ctx context.Context, firstName string) ([]model.User, error)

	// GetAllByFirstNameAndLastNameAndRole returns users filtered by full name and role constraints.
	GetAllByFirstNameAndLastNameAndRole(ctx context.Context, firstName, lastName string) ([]model.User, error)

	// GetByVkId returns a user by VK identifier.
	GetByVkId(ctx context.Context, vk int64) (*model.User, error)

	// Create stores a new user.
	Create(ctx context.Context, user *model.User) error

	// Update updates a single user field identified by VK identifier.
	Update(ctx context.Context, vkId int64, field string, value interface{}) (bool, error)

	// Delete removes a user by identifier.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistEventOrg reports whether the user is assigned as an event organizer anywhere.
	ExistEventOrg(ctx context.Context, userId uuid.UUID) (bool, error)
}
