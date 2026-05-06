package impl

import (
	"context"
	"ne_noy/internal/apperror"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/repository"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventParticipantRepository struct {
	pool *pgxpool.Pool
}

func NewEventParticipantRepository(pool *pgxpool.Pool) repository.EventParticipantRepository {
	return &eventParticipantRepository{pool: pool}
}

func (er *eventParticipantRepository) CheckParticipant(ctx context.Context, participant *as_event.EventParticipants) error {
	var exists bool
	row := er.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM event_participants WHERE event_id = $1 AND user_id = $2)`, participant.EventID, participant.UserID)
	if err := row.Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return apperror.ParticipantNotExistErr
	}

	_, err := er.pool.Exec(ctx, `
		UPDATE event_participants SET is_checked = $1, check_timestamp = $2, check_lat = $3, check_lon = $4, check_type = $5, check_author = $6
		WHERE event_id = $7 AND user_id = $8
	`, participant.IsChecked, participant.CheckTimestamp, participant.CheckLat, participant.CheckLong, participant.CheckType, participant.CheckAuthor, participant.EventID, participant.UserID)
	return err
}

func (er *eventParticipantRepository) Participant(ctx context.Context, eventId uuid.UUID, userVkId int64, prepareType string) (bool, error) {
	// Получить id пользователя и роль
	var userId uuid.UUID
	var roleId *uuid.UUID
	row := er.pool.QueryRow(ctx, `SELECT u.id, u.role_id FROM users u WHERE u.vk_id = $1`, userVkId)
	if err := row.Scan(&userId, &roleId); err != nil {
		return false, err
	}

	// Получить available roles для события
	rows, err := er.pool.Query(ctx, `SELECT role_id FROM event_roles WHERE event_id = $1`, eventId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	available := make([]uuid.UUID, 0)
	for rows.Next() {
		var r uuid.UUID
		if err := rows.Scan(&r); err != nil {
			return false, err
		}
		available = append(available, r)
	}

	// Если у пользователя нет роли — считаем, что не разрешено
	if roleId == nil {
		return false, apperror.UserRoleNotInAvailableRolesErr
	}

	// Проверяем, что роль пользователя есть в available
	if !slices.Contains(available, *roleId) {
		return false, apperror.UserRoleNotInAvailableRolesErr
	}

	// Вставляем запись участника
	_, err = er.pool.Exec(ctx, `INSERT INTO event_participants (event_id, user_id, prepare_type, created_at) VALUES ($1,$2,$3,now())`, eventId, userId, prepareType)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (er *eventParticipantRepository) ParticipantById(ctx context.Context, eventId uuid.UUID, userId uuid.UUID, prepareType string) (bool, error) {
	// Получаем роли события
	rows, err := er.pool.Query(ctx, `SELECT role_id FROM event_roles WHERE event_id = $1`, eventId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	available := make([]uuid.UUID, 0)
	for rows.Next() {
		var r uuid.UUID
		if err := rows.Scan(&r); err != nil {
			return false, err
		}
		available = append(available, r)
	}

	// Получаем роль пользователя
	var roleId uuid.UUID
	row := er.pool.QueryRow(ctx, `SELECT role_id FROM users WHERE id = $1`, userId)
	if err := row.Scan(&roleId); err != nil {
		return false, err
	}

	if !slices.Contains(available, roleId) {
		return false, apperror.UserRoleNotInAvailableRolesErr
	}

	_, err = er.pool.Exec(ctx, `INSERT INTO event_participants (event_id, user_id, prepare_type, created_at) VALUES ($1,$2,$3,now())`, eventId, userId, prepareType)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (er *eventParticipantRepository) UnParticipant(ctx context.Context, eventId uuid.UUID, userId int64) (bool, error) {
	// Удаляем по подзапросу: получить id участника по vk_id
	subRows, err := er.pool.Query(ctx, `
		SELECT ep.id FROM event_participants ep
		INNER JOIN users u ON ep.user_id = u.id
		WHERE ep.event_id = $1 AND u.vk_id = $2
	`, eventId, userId)
	if err != nil {
		return false, err
	}
	defer subRows.Close()

	ids := make([]uuid.UUID, 0)
	for subRows.Next() {
		var id uuid.UUID
		if err := subRows.Scan(&id); err != nil {
			return false, err
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return false, nil
	}

	// Удаляем записи
	for _, id := range ids {
		if _, err := er.pool.Exec(ctx, `DELETE FROM event_participants WHERE id = $1`, id); err != nil {
			return false, err
		}
	}
	return true, nil
}
