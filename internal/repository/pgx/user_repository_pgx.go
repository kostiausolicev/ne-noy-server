package pgx

import (
	"context"
	"errors"
	"fmt"
	"ne_noy/internal/config"
	"ne_noy/internal/model"
	"ne_noy/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{pool: pool}
}

func (u *userRepository) GetAll(ctx context.Context) ([]model.User, error) {
	rows, err := u.pool.Query(ctx, `
		SELECT 
			u.id, u.first_name, u.last_name, u.geo_available, u.notification_available, u.vk_id, u.photo_url,
		    r.id, r.name, r.display_name
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
	`)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	users := make([]model.User, 0)
	for rows.Next() {
		var role model.Role
		var user model.User

		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.GeoAvailable, &user.NotificationAvailable,
			&user.VkID, &user.PhotoURL, &role.ID, &role.Name, &role.DisplayName)
		if err != nil {
			return nil, err
		}
		user.Role = &role
		users = append(users, user)
	}
	return users, nil
}

func (u *userRepository) GetAllByFirstNameAndRole(ctx context.Context, firstName string) ([]model.User, error) {
	rows, err := u.pool.Query(ctx, `
		SELECT 
			u.id, u.first_name, u.last_name, u.geo_available, u.notification_available, u.vk_id, u.photo_url,
		    r.id, r.name, r.display_name
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE 1=1
			AND (u.first_name ILIKE $1 OR u.last_name ILIKE $1)
			AND r.name IN ($2, $3)
	`, fmt.Sprintf("%%%s%%", firstName), config.RoleHikePart, config.RoleAdmin)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	users := make([]model.User, 0)
	for rows.Next() {
		var role model.Role
		var user model.User

		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.GeoAvailable, &user.NotificationAvailable,
			&user.VkID, &user.PhotoURL, &role.ID, &role.Name, &role.DisplayName)
		if err != nil {
			return nil, err
		}
		user.Role = &role
		users = append(users, user)
	}
	return users, nil
}

func (u *userRepository) GetAllByFirstNameAndLastNameAndRole(ctx context.Context, firstName, lastName string) ([]model.User, error) {
	rows, err := u.pool.Query(ctx, `
		SELECT 
			u.id, u.first_name, u.last_name, u.geo_available, u.notification_available, u.vk_id, u.photo_url,
		    r.id, r.name, r.display_name
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE 1=1
			AND u.first_name ILIKE $1
			AND u.last_name ILIKE $2
			AND r.name IN ($3, $4)
	`, fmt.Sprintf("%%%s%%", firstName), fmt.Sprintf("%%%s%%", lastName), config.RoleHikePart, config.RoleAdmin)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	users := make([]model.User, 0)
	for rows.Next() {
		var role model.Role
		var user model.User

		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.GeoAvailable, &user.NotificationAvailable,
			&user.VkID, &user.PhotoURL, &role.ID, &role.Name, &role.DisplayName)
		if err != nil {
			return nil, err
		}
		user.Role = &role
		users = append(users, user)
	}
	return users, nil
}

func (u *userRepository) GetByVkId(ctx context.Context, vk int64) (*model.User, error) {
	var role model.Role
	var user model.User

	row := u.pool.QueryRow(ctx, `
		SELECT 
		    u.id, u.first_name, u.last_name, u.geo_available, u.notification_available, u.vk_id, u.photo_url,
		    r.id, r.name, r.display_name
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE u.vk_id = $1
	`, vk)

	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.GeoAvailable, &user.NotificationAvailable, &user.VkID, &user.PhotoURL, &role.ID, &role.Name, &role.DisplayName)
	if err != nil {
		return nil, err
	}
	user.Role = &role
	return &user, nil
}

func (u *userRepository) Create(ctx context.Context, user *model.User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	_, err := u.pool.Exec(ctx, `
		INSERT INTO users (id, vk_id, first_name, last_name, role_id, photo_url, geo_available, notification_available, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, user.ID, user.VkID, user.FirstName, user.LastName, user.RoleID, user.PhotoURL, user.GeoAvailable, user.NotificationAvailable, user.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (u *userRepository) Update(ctx context.Context, vkId int64, field string, value interface{}) (bool, error) {
	var query string
	switch field {
	case "role_id":
		query = "UPDATE users SET role_id = $1 WHERE vk_id = $2"
	case "geo_available":
		query = "UPDATE users SET geo_available = $1 WHERE vk_id = $2"
	case "notification_available":
		query = "UPDATE users SET notification_available = $1 WHERE vk_id = $2"
	default:
		return false, errors.New("invalid field")
	}

	// Выполняем обновление
	tag, err := u.pool.Exec(ctx, query, value, vkId)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (u *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := u.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if err != nil {
		return err
	}
	return nil
}

func (u *userRepository) ExistEventOrg(ctx context.Context, userId uuid.UUID) (bool, error) {
	var exists bool
	row := u.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM event_orgs
			WHERE user_id = $1
		)
	`, userId)
	err := row.Scan(&exists)
	return exists, err
}
