package impl

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type roleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepositoryPgx(poole *pgxpool.Pool) repository.RoleRepository {
	return &roleRepository{pool: poole}
}

func (r *roleRepository) GetById(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	role := model.Role{}
	row := r.pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.display_name FROM roles r
		WHERE r.id = $1
	`, id)
	err := row.Scan(&role.ID, &role.Name, &role.DisplayName)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	role := model.Role{}
	row := r.pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.display_name FROM roles r
		WHERE r.name = $1
	`, code)
	err := row.Scan(&role.ID, &role.Name, &role.DisplayName)
	if err != nil {
		return nil, err
	}
	return &role, nil
}
