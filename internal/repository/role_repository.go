package repository

import (
	"context"
	"ne_noy/internal/model"

	"github.com/google/uuid"
)

// RoleRepository defines methods for roles
type RoleRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
}
