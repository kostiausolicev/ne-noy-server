package impl

import (
	"context"
	"fmt"
	"ne_noy/internal/model/events"
	"ne_noy/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ne_noy/internal/model"
)

type eventRepositoryPgx struct {
	pool *pgxpool.Pool
}

func (e *eventRepositoryPgx) GetAll(ctx context.Context, roleCode *string, archived *bool) ([]*events.EventView, error) {
	// Собираем условие динамически, чтобы один метод покрывал и обычный список, и архив.
	conditions := make([]string, 0, 2)
	args := make([]interface{}, 0, 1)

	if archived == nil || !*archived {
		conditions = append(conditions, "(ends_at IS NULL OR ends_at >= NOW())")
	} else {
		conditions = append(conditions, "ends_at < NOW()")
	}

	if roleCode != nil {
		args = append(args, *roleCode)
		conditions = append(conditions, fmt.Sprintf(`
			EXISTS (
				SELECT 1
				FROM event_roles er
				INNER JOIN roles r ON r.id = er.role_id
				WHERE er.event_id = e.id
					AND er.event_type = e.type::event_type_enum
					AND r.name = $%d
			)
		`, len(args)))
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, e.type
		FROM events e
		WHERE e.status = 'ACTIVE' AND %s
		ORDER BY e.starts_at ASC
	`, strings.Join(conditions, " AND "))

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*events.EventView, 0)
	for rows.Next() {
		eventView, scanErr := scanEventView(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		// Для карточек мероприятий сразу подгружаем организаторов, чтобы верхний слой не делал N+1 вручную.
		orgs, orgErr := e.getEventOrgs(ctx, eventView.ID, eventView.Type, 3)
		if orgErr != nil {
			return nil, orgErr
		}

		participants, participantsErr := e.getEventParticipants(ctx, eventView.ID, eventView.Type, 3)
		if participantsErr != nil {
			return nil, participantsErr
		}
		count, countErr := e.getEventParticipantsCount(ctx, eventView.ID, eventView.Type)
		if countErr != nil {
			return nil, countErr
		}
		eventView.Orgs = orgs
		eventView.Participants = participants
		eventView.ParticipantsCount = count
		result = append(result, eventView)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (e *eventRepositoryPgx) GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*events.EventView, error) {
	// Возвращаем все мероприятия, в которых пользователь записан организатором независимо от типа профиля.
	rows, err := e.pool.Query(ctx, `
		SELECT ev.id, ev.name, ev.status, ev.starts_at, ev.ends_at, ev.type
		FROM events ev
		INNER JOIN event_orgs eo ON eo.event_id = ev.id AND eo.event_type = ev.type
		WHERE eo.user_id = $1
		ORDER BY ev.starts_at ASC
	`, orgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*events.EventView, 0)
	for rows.Next() {
		eventView, scanErr := scanEventView(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		orgs, orgErr := e.getEventOrgs(ctx, eventView.ID, eventView.Type, 3)
		if orgErr != nil {
			return nil, orgErr
		}
		participants, participantsErr := e.getEventParticipants(ctx, eventView.ID, eventView.Type, 3)
		if participantsErr != nil {
			return nil, participantsErr
		}
		count, countErr := e.getEventParticipantsCount(ctx, eventView.ID, eventView.Type)
		if countErr != nil {
			return nil, countErr
		}
		eventView.Orgs = orgs
		eventView.Participants = participants
		eventView.ParticipantsCount = count
		result = append(result, eventView)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (e *eventRepositoryPgx) GetLocationById(ctx context.Context, id uuid.UUID) (lat, long *float64, err error) {
	// Координаты есть только у обычных и командных мероприятий, поэтому ищем запись в обоих профилях.
	row := e.pool.QueryRow(ctx, `
		SELECT lat, lon
		FROM (
			SELECT id, lat, lon FROM event_as_events
			UNION ALL
			SELECT id, lat, lon FROM event_as_teams
		) event_locations
		WHERE id = $1
		LIMIT 1
	`, id)

	if err = row.Scan(&lat, &long); err != nil {
		return nil, nil, err
	}

	return lat, long, nil
}

func (e *eventRepositoryPgx) Delete(ctx context.Context, id uuid.UUID, eventType string) error {
	// Удаляем запись из профильной таблицы по типу мероприятия, чтобы не затронуть другие профили.
	tableName, err := events.GetEventTableName(eventType)
	if err != nil {
		return err
	}

	commandTag, err := e.pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName), id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (e *eventRepositoryPgx) Publish(ctx context.Context, id uuid.UUID) error {
	var eventType string
	err := e.pool.QueryRow(ctx, `SELECT type FROM events WHERE id = $1`, id).Scan(&eventType)
	if err != nil {
		return err
	}

	tableName, err := events.GetEventTableName(eventType)
	if err != nil {
		return err
	}

	tag, err := e.pool.Exec(ctx, fmt.Sprintf(`UPDATE %s SET status = 'ACTIVE' WHERE id = $1`, tableName), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func NewEventBaseRepository(pool *pgxpool.Pool) repository.EventBaseRepository {
	return &eventRepositoryPgx{pool: pool}
}

func scanEventView(row pgx.Row) (*events.EventView, error) {
	var (
		eventView events.EventView
		startsAt  time.Time
		endsAt    *time.Time
	)

	if err := row.Scan(&eventView.ID, &eventView.Name, &eventView.Status, &startsAt, &endsAt, &eventView.Type); err != nil {
		return nil, err
	}

	eventView.StartsAt = startsAt
	eventView.EndsAt = endsAt
	return &eventView, nil
}

func (e *eventRepositoryPgx) getEventOrgs(ctx context.Context, id uuid.UUID, eventType string, limit int) ([]model.User, error) {
	// Один хелпер для выборки организаторов нужен всем типам мероприятий, различается только event_type.
	query := `
		SELECT
			u.id, u.created_at,
			u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			u.geo_available, u.notification_available
		FROM event_orgs eo
		INNER JOIN users u ON u.id = eo.user_id
		WHERE eo.event_id = $1 AND eo.event_type = $2::event_type_enum
		ORDER BY u.created_at ASC
	`
	args := []interface{}{id, eventType}
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orgs := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err = rows.Scan(
			&user.ID, &user.CreatedAt,
			&user.VkID, &user.FirstName, &user.LastName, &user.RoleID, &user.PhotoURL,
			&user.GeoAvailable, &user.NotificationAvailable,
		); err != nil {
			return nil, err
		}
		orgs = append(orgs, user)
	}

	return orgs, rows.Err()
}

// getEventParticipants возвращает превью участников мероприятия в зависимости от его типа:
//   - event  → записавшиеся участники из event_participants
//   - test   → пользователи, начавшие тест (имеющие хотя бы один ответ)
//   - team   → капитаны команд
func (e *eventRepositoryPgx) getEventParticipants(ctx context.Context, id uuid.UUID, eventType string, limit int) ([]model.User, error) {
	var query string
	args := []interface{}{id}

	switch eventType {
	case events.EventAsEvent:
		query = `
			SELECT u.id, u.created_at, u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			       u.geo_available, u.notification_available
			FROM event_participants ep
			INNER JOIN users u ON u.id = ep.user_id
			WHERE ep.event_id = $1
			ORDER BY ep.created_at ASC
		`
	case events.EventAsTest:
		query = `
			SELECT u.id, u.created_at, u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			       u.geo_available, u.notification_available
			FROM users u
			WHERE u.id IN (
				SELECT DISTINCT ua.user_id
				FROM user_answers ua
				INNER JOIN questions q ON q.id = ua.question_id
				WHERE q.event_id = $1
			)
			ORDER BY u.created_at ASC
		`
	case events.EventAsTeam:
		query = `
			SELECT u.id, u.created_at, u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			       u.geo_available, u.notification_available
			FROM teams t
			INNER JOIN users u ON u.id = t.captain_id
			WHERE t.event_id = $1
			ORDER BY t.created_at ASC
		`
	default:
		return nil, nil
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err = rows.Scan(
			&user.ID, &user.CreatedAt, &user.VkID, &user.FirstName, &user.LastName,
			&user.RoleID, &user.PhotoURL, &user.GeoAvailable, &user.NotificationAvailable,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// getEventParticipantsCount возвращает полное число участников мероприятия:
//   - event  → все записавшиеся участники
//   - test   → число уникальных пользователей, начавших тест
//   - team   → число команд (равно числу капитанов)
func (e *eventRepositoryPgx) getEventParticipantsCount(ctx context.Context, id uuid.UUID, eventType string) (int, error) {
	var count int
	var err error

	switch eventType {
	case events.EventAsEvent:
		err = e.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM event_participants WHERE event_id = $1
		`, id).Scan(&count)
	case events.EventAsTest:
		err = e.pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT ua.user_id)
			FROM user_answers ua
			INNER JOIN questions q ON q.id = ua.question_id
			WHERE q.event_id = $1
		`, id).Scan(&count)
	case events.EventAsTeam:
		err = e.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM teams WHERE event_id = $1
		`, id).Scan(&count)
	default:
		return 0, nil
	}

	return count, err
}
