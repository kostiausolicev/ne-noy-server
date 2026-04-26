package pgx

import (
	"context"
	"fmt"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/model/events/as_team"
	"ne_noy/internal/model/events/as_test"
	"ne_noy/internal/repository"
	"slices"
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
					AND er.event_type = e.type
					AND r.name = $%d
			)
		`, len(args)))
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, e.type
		FROM events e
		WHERE %s
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
		orgs, orgErr := e.getEventOrgsByType(ctx, eventView.ID, eventView.Type, 3)
		if orgErr != nil {
			return nil, orgErr
		}
		eventView.Orgs = orgs
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

		orgs, orgErr := e.getEventOrgsByType(ctx, eventView.ID, eventView.Type, 3)
		if orgErr != nil {
			return nil, orgErr
		}
		eventView.Orgs = orgs
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

func (e *eventRepositoryPgx) ExistUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) (bool, error) {
	// Проверяем наличие записи участника напрямую через EXISTS, чтобы не тянуть лишние данные.
	var exists bool
	row := e.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM event_participants
			WHERE event_id = $1 AND user_id = $2
		)
	`, eventId, userId)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (e *eventRepositoryPgx) GetEventOrgs(ctx context.Context, id uuid.UUID, limit int) ([]model.User, error) {
	// Сначала узнаём тип мероприятия, потому что связи организаторов разделены по event_type.
	eventType, err := e.getEventTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return e.getEventOrgsByType(ctx, id, eventType, limit)
}

func (e *eventRepositoryPgx) GetByVkPollId(ctx context.Context, pollId int64) (*as_event.AsEvent, error) {
	// Для обычных мероприятий идентификатор опроса хранится в vk_vote_id, поэтому ищем профиль по нему.
	return e.getAsEventByCondition(ctx, "vk_vote_id = $1", pollId)
}

func (e *eventRepositoryPgx) GetEventById(ctx context.Context, id uuid.UUID) (*as_event.AsEvent, error) {
	// Для загрузки обычного мероприятия по идентификатору используем тот же путь, что и для поиска по vk-полям.
	return e.getAsEventByCondition(ctx, "id = $1", id)
}

func (e *eventRepositoryPgx) GetTeamById(ctx context.Context, id uuid.UUID) (*as_team.AsTeam, error) {
	// Командное мероприятие загружаем вместе с организаторами, вложениями и количеством участников.
	return e.getAsTeamByCondition(ctx, "id = $1", id)
}

func (e *eventRepositoryPgx) GetTestById(ctx context.Context, id uuid.UUID) (*as_test.AsTest, error) {
	// Тест загружаем вместе со структурой вопросов, потому что без неё профиль теста практически бесполезен.
	return e.getAsTestByCondition(ctx, "id = $1", id)
}

func (e *eventRepositoryPgx) GetParticipants(ctx context.Context, id uuid.UUID, limit int) ([]as_event.EventParticipant, error) {
	// Получаем участников вместе с данными пользователя, чтобы вызывающему коду не пришлось собирать их вручную.
	query := `
		SELECT
			ep.id, ep.created_at, ep.event_id, ep.user_id, ep.prepare_type,
			ep.is_checked, ep.check_timestamp, ep.check_lat, ep.check_lon, ep.check_type, ep.check_author,
			u.id, u.created_at, u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			u.geo_available, u.notification_available
		FROM event_participants ep
		INNER JOIN users u ON u.id = ep.user_id
		WHERE ep.event_id = $1
		ORDER BY ep.created_at ASC
	`
	args := []interface{}{id}
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := make([]as_event.EventParticipant, 0)
	for rows.Next() {
		var participant as_event.EventParticipant
		if err = rows.Scan(
			&participant.ID, &participant.CreatedAt, &participant.EventID, &participant.UserID, &participant.PrepareType,
			&participant.IsChecked, &participant.CheckTimestamp, &participant.CheckLat, &participant.CheckLong, &participant.CheckType, &participant.CheckAuthor,
			&participant.User.ID, &participant.User.CreatedAt, &participant.User.VkID, &participant.User.FirstName, &participant.User.LastName, &participant.User.RoleID, &participant.User.PhotoURL,
			&participant.User.GeoAvailable, &participant.User.NotificationAvailable,
		); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	return participants, rows.Err()
}

func (e *eventRepositoryPgx) CreateEvent(ctx context.Context, event *as_event.AsEvent) (*as_event.AsEvent, error) {
	// Создаём обычное мероприятие и его связи в одной транзакции, чтобы состояние не расходилось.
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO event_as_events (
			id, name, description, cover, status, starts_at, ends_at,
			vk_post_id, vk_vote_id, lat, lon, address, additional_address, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		)
	`,
		event.ID, event.Name, event.Description, event.Cover, event.Status, event.StartsAt, event.EndsAt,
		event.VkPostID, event.VkVoteID, event.Lat, event.Lon, event.Address, event.AdditionalAddress, event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = e.insertEventOrgs(ctx, tx, event.ID, events.EventAsEvent, event.Orgs); err != nil {
		return nil, err
	}
	if err = e.insertEventAttachments(ctx, tx, event.ID, events.EventAsEvent, event.Attachments); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e.GetEventById(ctx, event.ID)
}

func (e *eventRepositoryPgx) CreateTeam(ctx context.Context, team *as_team.AsTeam) (*as_team.AsTeam, error) {
	// Командное мероприятие создаём атомарно: профиль, организаторы и вложения должны сохраниться вместе.
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO event_as_teams (
			id, name, description, cover, status, starts_at, ends_at,
			teams_constraint, teams_cap_min, teams_cap_max,
			lat, lon, address, additional_address, vk_post_id,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17
		)
	`,
		team.ID, team.Name, team.Description, team.Cover, team.Status, team.StartsAt, team.EndsAt,
		team.TeamsConstraint, team.TeamsCapMin, team.TeamsCapMax,
		team.Lat, team.Lon, team.Address, team.AdditionalAddress, team.VkPostID,
		team.CreatedAt, team.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = e.insertEventOrgs(ctx, tx, team.ID, events.EventAsTeam, team.Orgs); err != nil {
		return nil, err
	}
	if err = e.insertEventAttachments(ctx, tx, team.ID, events.EventAsTeam, team.Attachments); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e.GetTeamById(ctx, team.ID)
}

func (e *eventRepositoryPgx) CreateTest(ctx context.Context, test *as_test.AsTest) (*as_test.AsTest, error) {
	// Тест создаём транзакционно, потому что вопросы и ответы должны соответствовать одной версии профиля.
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO event_as_tests (
			id, name, description, cover, status, starts_at, ends_at,
			ext_link_id, attempts, vk_post_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12
		)
	`,
		test.ID, test.Name, test.Description, test.Cover, test.Status, test.StartsAt, test.EndsAt,
		test.ExtLinkID, test.Attempts, test.VkPostID, test.CreatedAt, test.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = e.insertEventOrgs(ctx, tx, test.ID, events.EventAsTest, test.Orgs); err != nil {
		return nil, err
	}
	if err = e.insertEventAttachments(ctx, tx, test.ID, events.EventAsTest, test.Attachments); err != nil {
		return nil, err
	}
	if err = e.insertTestQuestions(ctx, tx, test.ID, test.Questions); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e.GetTestById(ctx, test.ID)
}

func (e *eventRepositoryPgx) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*events.EventView, error) {
	// Обновление выполняем в транзакции, чтобы профиль и его связи не расходились между собой.
	eventType, err := e.getEventTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tableName, err := getEventTableName(eventType)
	if err != nil {
		return nil, err
	}

	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if len(fields) > 0 {
		setParts := make([]string, 0, len(fields))
		args := make([]interface{}, 0, len(fields)+1)
		position := 1

		for _, column := range getSortedAllowedColumns(fields) {
			args = append(args, dereferenceValue(fields[column]))
			setParts = append(setParts, fmt.Sprintf("%s = $%d", column, position))
			position++
		}

		args = append(args, id)
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, strings.Join(setParts, ", "), position)
		if _, err = tx.Exec(ctx, query, args...); err != nil {
			return nil, err
		}
	}

	if orgs != nil {
		// Организаторов при update заменяем полностью, чтобы состояние отражало переданный набор.
		if _, err = tx.Exec(ctx, `DELETE FROM event_orgs WHERE event_id = $1 AND event_type = $2`, id, eventType); err != nil {
			return nil, err
		}
		if err = e.insertEventOrgs(ctx, tx, id, eventType, orgs); err != nil {
			return nil, err
		}
	}

	if availableRoles != nil {
		// Доступные роли тоже пересоздаём целиком, иначе сложно корректно синхронизировать удалённые элементы.
		if _, err = tx.Exec(ctx, `DELETE FROM event_roles WHERE event_id = $1 AND event_type = $2`, id, eventType); err != nil {
			return nil, err
		}
		if err = e.insertEventRoles(ctx, tx, id, eventType, availableRoles); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return e.getEventViewByID(ctx, id)
}

func (e *eventRepositoryPgx) Delete(ctx context.Context, id uuid.UUID, eventType string) error {
	// Удаляем запись из профильной таблицы по типу мероприятия, чтобы не затронуть другие профили.
	tableName, err := getEventTableName(eventType)
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

func NewEventRepositoryPgx(pool *pgxpool.Pool) repository.EventRepository {
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

func (e *eventRepositoryPgx) getEventOrgsByType(ctx context.Context, id uuid.UUID, eventType string, limit int) ([]model.User, error) {
	// Один хелпер для выборки организаторов нужен всем типам мероприятий, различается только event_type.
	query := `
		SELECT
			u.id, u.created_at,
			u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			u.geo_available, u.notification_available
		FROM event_orgs eo
		INNER JOIN users u ON u.id = eo.user_id
		WHERE eo.event_id = $1 AND eo.event_type = $2
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

func (e *eventRepositoryPgx) getEventTypeByID(ctx context.Context, id uuid.UUID) (string, error) {
	// Тип берём из общей вью, чтобы не повторять проверку по каждой профильной таблице вручную.
	var eventType string
	row := e.pool.QueryRow(ctx, `SELECT type FROM events WHERE id = $1`, id)
	if err := row.Scan(&eventType); err != nil {
		return "", err
	}

	return eventType, nil
}

func (e *eventRepositoryPgx) getAsEventByCondition(ctx context.Context, condition string, args ...interface{}) (*as_event.AsEvent, error) {
	// Базовый профиль обычного мероприятия читаем одним запросом, а связи добираем отдельными хелперами.
	query := fmt.Sprintf(`
		SELECT
			id, created_at, name, description, cover, status,
			starts_at, ends_at, vk_post_id, vk_vote_id, lat, lon, address, additional_address
		FROM event_as_events
		WHERE %s
		LIMIT 1
	`, condition)

	row := e.pool.QueryRow(ctx, query, args...)

	var event as_event.AsEvent
	if err := row.Scan(
		&event.ID, &event.CreatedAt, &event.Name, &event.Description, &event.Cover, &event.Status,
		&event.StartsAt, &event.EndsAt, &event.VkPostID, &event.VkVoteID, &event.Lat, &event.Lon, &event.Address, &event.AdditionalAddress,
	); err != nil {
		return nil, err
	}

	orgs, err := e.getEventOrgsByType(ctx, event.ID, events.EventAsEvent, 0)
	if err != nil {
		return nil, err
	}
	event.Orgs = orgs

	attachments, err := e.getEventAttachmentsByType(ctx, event.ID, events.EventAsEvent)
	if err != nil {
		return nil, err
	}
	event.Attachments = attachments

	participantsCount, err := e.getEventParticipantsCount(ctx, event.ID)
	if err != nil {
		return nil, err
	}
	event.ParticipantsCount = participantsCount

	return &event, nil
}

func (e *eventRepositoryPgx) getEventAttachmentsByType(ctx context.Context, id uuid.UUID, eventType string) ([]events.EventAttachment, error) {
	// Вложения подтягиваем через таблицу связей, чтобы вернуть и метаданные attachment, и технический id связи.
	rows, err := e.pool.Query(ctx, `
		SELECT
			ea.id, ea.created_at, ea.event_id, ea.event_type, ea.attachment_id,
			a.id, a.filename, a.url, a.created_at
		FROM event_attachments ea
		INNER JOIN attachments a ON a.id = ea.attachment_id
		WHERE ea.event_id = $1 AND ea.event_type = $2
		ORDER BY ea.created_at ASC
	`, id, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]events.EventAttachment, 0)
	for rows.Next() {
		var attachment events.EventAttachment
		if err = rows.Scan(
			&attachment.ID, &attachment.CreatedAt, &attachment.EventID, &attachment.EventType, &attachment.AttachmentID,
			&attachment.Attachment.ID, &attachment.Attachment.Filename, &attachment.Attachment.Url, &attachment.Attachment.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, rows.Err()
}

func (e *eventRepositoryPgx) getEventParticipantsCount(ctx context.Context, id uuid.UUID) (int, error) {
	// Количество участников хранится агрегированно в event_participants, поэтому считаем его отдельно.
	var count int
	row := e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM event_participants WHERE event_id = $1`, id)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (e *eventRepositoryPgx) getAsTeamByCondition(ctx context.Context, condition string, args ...interface{}) (*as_team.AsTeam, error) {
	// Базовые поля командного мероприятия читаем из профильной таблицы, а связанные сущности добираем отдельно.
	query := fmt.Sprintf(`
		SELECT
			id, created_at, updated_at, name, description, cover, status,
			starts_at, ends_at, teams_constraint, teams_cap_min, teams_cap_max,
			lat, lon, address, additional_address, vk_post_id
		FROM event_as_teams
		WHERE %s
		LIMIT 1
	`, condition)

	row := e.pool.QueryRow(ctx, query, args...)

	var team as_team.AsTeam
	if err := row.Scan(
		&team.ID, &team.CreatedAt, &team.UpdatedAt, &team.Name, &team.Description, &team.Cover, &team.Status,
		&team.StartsAt, &team.EndsAt, &team.TeamsConstraint, &team.TeamsCapMin, &team.TeamsCapMax,
		&team.Lat, &team.Lon, &team.Address, &team.AdditionalAddress, &team.VkPostID,
	); err != nil {
		return nil, err
	}

	orgs, err := e.getEventOrgsByType(ctx, team.ID, events.EventAsTeam, 0)
	if err != nil {
		return nil, err
	}
	team.Orgs = orgs

	attachments, err := e.getEventAttachmentsByType(ctx, team.ID, events.EventAsTeam)
	if err != nil {
		return nil, err
	}
	team.Attachments = attachments

	return &team, nil
}

func (e *eventRepositoryPgx) getAsTestByCondition(ctx context.Context, condition string, args ...interface{}) (*as_test.AsTest, error) {
	// Основную запись теста читаем отдельно, а затем подгружаем связи и набор вопросов.
	query := fmt.Sprintf(`
		SELECT
			id, created_at, updated_at, name, description, cover, status,
			starts_at, ends_at, ext_link_id, attempts, vk_post_id
		FROM event_as_tests
		WHERE %s
		LIMIT 1
	`, condition)

	row := e.pool.QueryRow(ctx, query, args...)

	var test as_test.AsTest
	if err := row.Scan(
		&test.ID, &test.CreatedAt, &test.UpdatedAt, &test.Name, &test.Description, &test.Cover, &test.Status,
		&test.StartsAt, &test.EndsAt, &test.ExtLinkID, &test.Attempts, &test.VkPostID,
	); err != nil {
		return nil, err
	}

	orgs, err := e.getEventOrgsByType(ctx, test.ID, events.EventAsTest, 0)
	if err != nil {
		return nil, err
	}
	test.Orgs = orgs

	attachments, err := e.getEventAttachmentsByType(ctx, test.ID, events.EventAsTest)
	if err != nil {
		return nil, err
	}
	test.Attachments = attachments

	questions, err := e.getTestQuestions(ctx, test.ID)
	if err != nil {
		return nil, err
	}
	test.Questions = questions

	return &test, nil
}

func (e *eventRepositoryPgx) getTestQuestions(ctx context.Context, eventID uuid.UUID) ([]*as_test.Question, error) {
	// Вопросы собираем по порядку, а ответы и вложения для каждого вопроса добираем отдельными хелперами.
	rows, err := e.pool.Query(ctx, `
		SELECT id, created_at, updated_at, text, type, event_id, q_order
		FROM questions
		WHERE event_id = $1
		ORDER BY q_order ASC, created_at ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := make([]*as_test.Question, 0)
	for rows.Next() {
		question := &as_test.Question{}
		if err = rows.Scan(
			&question.ID, &question.CreatedAt, &question.UpdatedAt, &question.Text, &question.Type, &question.EventID, &question.QOrder,
		); err != nil {
			return nil, err
		}

		answers, answersErr := e.getQuestionAnswers(ctx, question.ID)
		if answersErr != nil {
			return nil, answersErr
		}
		question.Answers = answers

		attachments, attachmentsErr := e.getQuestionAttachments(ctx, question.ID)
		if attachmentsErr != nil {
			return nil, attachmentsErr
		}
		question.Attachments = attachments

		questions = append(questions, question)
	}

	return questions, rows.Err()
}

func (e *eventRepositoryPgx) getQuestionAnswers(ctx context.Context, questionID uuid.UUID) ([]*as_test.Answer, error) {
	// Варианты ответов возвращаем в том виде, в котором они лежат в таблице answers.
	rows, err := e.pool.Query(ctx, `
		SELECT id, created_at, updated_at, question_id, is_correct, text, points
		FROM answers
		WHERE question_id = $1
		ORDER BY created_at ASC
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := make([]*as_test.Answer, 0)
	for rows.Next() {
		answer := &as_test.Answer{}
		if err = rows.Scan(
			&answer.ID, &answer.CreatedAt, &answer.UpdatedAt, &answer.QuestionID, &answer.IsCorrect, &answer.Text, &answer.Points,
		); err != nil {
			return nil, err
		}
		answers = append(answers, answer)
	}

	return answers, rows.Err()
}

func (e *eventRepositoryPgx) getQuestionAttachments(ctx context.Context, questionID uuid.UUID) ([]*as_test.QuestionAttachment, error) {
	// Вложения к вопросам возвращаем вместе с описанием файла, чтобы верхнему слою не пришлось делать допзапросы.
	rows, err := e.pool.Query(ctx, `
		SELECT
			qa.id, qa.created_at, qa.updated_at, qa.question_id,
			a.id, a.filename, a.url, a.created_at
		FROM question_attachments qa
		LEFT JOIN attachments a ON a.id = qa.attachment_id
		WHERE qa.question_id = $1
		ORDER BY qa.created_at ASC
	`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]*as_test.QuestionAttachment, 0)
	for rows.Next() {
		attachment := &as_test.QuestionAttachment{}
		if err = rows.Scan(
			&attachment.ID, &attachment.CreatedAt, &attachment.UpdatedAt, &attachment.QuestionId,
			&attachment.Attachment.ID, &attachment.Attachment.Filename, &attachment.Attachment.Url, &attachment.Attachment.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, rows.Err()
}

func (e *eventRepositoryPgx) insertEventOrgs(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, eventType string, orgs []model.User) error {
	// Организаторов вставляем отдельными строками, потому что связь many-to-many хранится в event_orgs.
	for _, org := range orgs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_orgs (event_id, user_id, event_type)
			VALUES ($1, $2, $3)
		`, eventID, org.ID, eventType); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventRepositoryPgx) insertEventAttachments(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, eventType string, attachments []events.EventAttachment) error {
	// Вложения события сохраняем через таблицу связей, используя уже существующий attachment_id.
	for _, attachment := range attachments {
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_attachments (id, event_id, event_type, attachment_id, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, attachment.ID, eventID, eventType, attachment.AttachmentID, attachment.CreatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventRepositoryPgx) insertTestQuestions(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, questions []*as_test.Question) error {
	// Вопросы и ответы вставляем последовательно, сохраняя их принадлежность конкретному тесту.
	for _, question := range questions {
		_, err := tx.Exec(ctx, `
			INSERT INTO questions (id, text, type, event_id, q_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, question.ID, question.Text, question.Type, eventID, question.QOrder, question.CreatedAt, question.UpdatedAt)
		if err != nil {
			return err
		}

		for _, answer := range question.Answers {
			if _, err = tx.Exec(ctx, `
				INSERT INTO answers (id, question_id, is_correct, text, points, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, answer.ID, question.ID, answer.IsCorrect, answer.Text, answer.Points, answer.CreatedAt, answer.UpdatedAt); err != nil {
				return err
			}
		}

		for _, attachment := range question.Attachments {
			if _, err = tx.Exec(ctx, `
				INSERT INTO question_attachments (id, question_id, attachment_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
			`, attachment.ID, question.ID, attachment.Attachment.ID, attachment.CreatedAt, attachment.UpdatedAt); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *eventRepositoryPgx) insertEventRoles(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, eventType string, roles []model.Role) error {
	// Доступность мероприятия по ролям хранится в отдельной таблице event_roles.
	for _, role := range roles {
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_roles (event_id, role_id, event_type)
			VALUES ($1, $2, $3)
		`, eventID, role.ID, eventType); err != nil {
			return err
		}
	}

	return nil
}

func getEventTableName(eventType string) (string, error) {
	switch eventType {
	case events.EventAsEvent:
		return "event_as_events", nil
	case events.EventAsTeam:
		return "event_as_teams", nil
	case events.EventAsTest:
		return "event_as_tests", nil
	default:
		return "", fmt.Errorf("unsupported event type: %s", eventType)
	}
}

func getSortedAllowedColumns(fields map[string]interface{}) []string {
	// Обновлять разрешаем только известные колонки, чтобы не допустить случайной подстановки произвольного SQL.
	allowed := map[string]struct{}{
		"name": {}, "description": {}, "cover": {}, "status": {}, "starts_at": {}, "ends_at": {},
		"vk_post_id": {}, "vk_vote_id": {}, "vk_poll_answer_id": {}, "lat": {}, "lon": {}, "address": {},
		"additional_address": {}, "ext_link_id": {}, "attempts": {}, "teams_constraint": {}, "teams_cap_min": {}, "teams_cap_max": {},
	}

	columns := make([]string, 0, len(fields))
	for column := range fields {
		if _, ok := allowed[column]; ok {
			columns = append(columns, column)
		}
	}
	slices.Sort(columns)
	return columns
}

func dereferenceValue(value interface{}) interface{} {
	// В map обновления иногда прилетают указатели, поэтому разворачиваем самые частые типы перед передачей в pgx.
	switch typed := value.(type) {
	case *string:
		if typed == nil {
			return nil
		}
		return *typed
	case *int:
		if typed == nil {
			return nil
		}
		return *typed
	case *int64:
		if typed == nil {
			return nil
		}
		return *typed
	case *float64:
		if typed == nil {
			return nil
		}
		return *typed
	case *time.Time:
		if typed == nil {
			return nil
		}
		return *typed
	default:
		return value
	}
}

func (e *eventRepositoryPgx) getEventViewByID(ctx context.Context, id uuid.UUID) (*events.EventView, error) {
	// После мутации возвращаем карточку мероприятия из общей вью, чтобы вызывающий код получил актуальный тип и даты.
	row := e.pool.QueryRow(ctx, `
		SELECT id, name, status, starts_at, ends_at, type
		FROM events
		WHERE id = $1
	`, id)

	eventView, err := scanEventView(row)
	if err != nil {
		return nil, err
	}

	orgs, err := e.getEventOrgsByType(ctx, eventView.ID, eventView.Type, 0)
	if err != nil {
		return nil, err
	}
	eventView.Orgs = orgs

	return eventView, nil
}
