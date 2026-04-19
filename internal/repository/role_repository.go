package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// RoleRepository describes persistence operations for roles.
type RoleRepository interface {
	// GetById returns a role by its identifier.
	GetById(ctx context.Context, id uuid.UUID) (*model.Role, error)

	// GetByCode returns a role by its code or name.
	GetByCode(ctx context.Context, code string) (*model.Role, error)
}
