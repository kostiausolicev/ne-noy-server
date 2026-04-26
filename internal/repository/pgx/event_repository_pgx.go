package pgx

import (
	"context"
	"errors"
	"fmt"
	"ne_noy/internal/model/events"
	"ne_noy/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ne_noy/internal/model"
)

type eventRepositoryPgx struct {
	pool *pgxpool.Pool
}

const defaultEventType = model.EventTypeEvent

func NewEventRepositoryPgx(pool *pgxpool.Pool) repository.EventRepository {
	return &eventRepositoryPgx{pool: pool}
}

func (r *eventRepositoryPgx) resolveEventType(ctx context.Context, id uuid.UUID) (string, error) {
	var eventType string
	if err := r.pool.QueryRow(ctx, `SELECT type FROM events WHERE id = $1`, id).Scan(&eventType); err != nil {
		return "", err
	}
	return eventType, nil
}

func (r *eventRepositoryPgx) GetEventTypeById(ctx context.Context, id uuid.UUID) (string, error) {
	return r.resolveEventType(ctx, id)
}

func (r *eventRepositoryPgx) resolveTableName(ctx context.Context, id uuid.UUID) (string, string, error) {
	eventType, err := r.resolveEventType(ctx, id)
	if err != nil {
		return "", "", err
	}

	switch eventType {
	case model.EventTypeEvent:
		return "event_as_events", eventType, nil
	case model.EventTypeActivity:
		return "event_as_activities", eventType, nil
	case model.EventTypeTeam:
		return "event_as_teams", eventType, nil
	case model.EventTypePoll:
		return "event_as_polls", eventType, nil
	case model.EventTypeTest:
		return "event_as_tests", eventType, nil
	default:
		return "", "", fmt.Errorf("unsupported as_event type: %s", eventType)
	}
}

func eventDetailSelect(tableName, eventType string) string {
	switch eventType {
	case model.EventTypeEvent:
		return fmt.Sprintf(`
			SELECT id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon
			FROM %s WHERE id = $1
		`, tableName)
	case model.EventTypeTeam:
		return fmt.Sprintf(`
			SELECT id, name, cover, description, address, additional_address, vk_post_id, NULL::bigint, NULL::bigint, status, starts_at, ends_at, lat, lon
			FROM %s WHERE id = $1
		`, tableName)
	case model.EventTypePoll:
		return fmt.Sprintf(`
			SELECT id, name, cover, description, NULL::text, NULL::varchar, vk_post_id, NULL::bigint, NULL::bigint, status, starts_at, ends_at, NULL::numeric, NULL::numeric
			FROM %s WHERE id = $1
		`, tableName)
	case model.EventTypeTest:
		return fmt.Sprintf(`
			SELECT id, name, cover, description, NULL::text, NULL::varchar, vk_post_id, NULL::bigint, NULL::bigint, status, starts_at, ends_at, NULL::numeric, NULL::numeric
			FROM %s WHERE id = $1
		`, tableName)
	case model.EventTypeActivity:
		return fmt.Sprintf(`
			SELECT id, name, cover, description, NULL::text, NULL::varchar, NULL::bigint, NULL::bigint, NULL::bigint, status, starts_at, ends_at, NULL::numeric, NULL::numeric
			FROM %s WHERE id = $1
		`, tableName)
	default:
		return ""
	}
}

func eventLocationSelect(tableName, eventType string) string {
	switch eventType {
	case model.EventTypeEvent, model.EventTypeTeam:
		return fmt.Sprintf(`SELECT id, lat, lon FROM %s WHERE id = $1`, tableName)
	default:
		return fmt.Sprintf(`SELECT id, NULL::numeric, NULL::numeric FROM %s WHERE id = $1`, tableName)
	}
}

func (r *eventRepositoryPgx) loadEventRelations(ctx context.Context, eventID uuid.UUID, eventType string) (model.EventRelations, error) {
	relations := model.EventRelations{}

	orgs, err := r.GetEventOrgs(ctx, eventID)
	if err != nil {
		return relations, err
	}
	relations.Orgs = orgs

	pRows, err := r.pool.Query(ctx, `
		SELECT ep.id, ep.user_id, ep.event_id, ep.is_checked, ep.check_timestamp, u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM event_participants ep
		JOIN users u ON u.id = ep.user_id
		WHERE ep.event_id = $1
		ORDER BY ep.created_at DESC
		LIMIT 3
	`, eventID)
	if err != nil {
		return relations, err
	}
	defer pRows.Close()

	for pRows.Next() {
		var ep model.EventParticipant
		var checkTimestamp *time.Time
		var u model.User
		if err := pRows.Scan(&ep.ID, &ep.UserID, &ep.EventID, &ep.IsChecked, &checkTimestamp, &u.ID, &u.VkID, &u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
			return relations, err
		}
		ep.CheckTimestamp = checkTimestamp
		ep.User = u
		relations.EventParticipants = append(relations.EventParticipants, ep)
	}

	attRows, err := r.pool.Query(ctx, `
		SELECT ea.id, a.id, a.url, a.filename
		FROM event_attachments ea
		JOIN attachments a ON a.id = ea.attachment_id
		WHERE ea.event_id = $1 AND (ea.event_type = $2 OR ea.event_type IS NULL)
	`, eventID, eventType)
	if err != nil {
		return relations, err
	}
	defer attRows.Close()

	for attRows.Next() {
		var a model.EventAttachment
		var attachID int64
		var url string
		var filename string
		if err := attRows.Scan(&a.ID, &attachID, &url, &filename); err != nil {
			return relations, err
		}
		a.EventType = eventType
		a.AttachmentID = attachID
		a.Attachment = model.Attachment{ID: attachID, Url: url, Filename: filename}
		relations.Attachments = append(relations.Attachments, a)
	}

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM event_participants WHERE event_id = $1`, eventID).Scan(&relations.ParticipantsCount); err != nil {
		return relations, err
	}

	return relations, nil
}

func (r *eventRepositoryPgx) GetAll(ctx context.Context) ([]*events.EventView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, status, (
				SELECT COUNT(*) 
				FROM event_participants ep 
				WHERE ep.event_id = events.id
			) AS participants_count, starts_at, type
		FROM events
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*events.EventView
	for rows.Next() {
		var e events.EventView
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt, &e.EventType); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*events.EventView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, e.status, (
				SELECT COUNT(*) 
				FROM event_participants ep 
				WHERE ep.event_id = e.id
			) AS participants_count, e.starts_at, e.type
		FROM events e
		LEFT JOIN event_orgs eo ON eo.event_id = e.id
		WHERE eo.user_id = $1 AND (eo.event_type = e.type OR eo.event_type IS NULL)
	`, orgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*events.EventView
	for rows.Next() {
		var e events.EventView
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt, &e.EventType); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetAllByRole(ctx context.Context, role string) ([]*events.EventView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, e.status, e.starts_at, (
				SELECT COUNT(*) 
				FROM event_participants ep 
				WHERE ep.event_id = e.id
			) AS participants_count, e.type
		FROM events e
		JOIN event_roles er ON er.event_id = e.id AND (er.event_type = e.type OR er.event_type IS NULL)
		JOIN roles r ON r.id = er.role_id
		WHERE r.name = $1 AND e.ends_at > NOW() AND e.status = 'ACTIVE'
	`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*events.EventView
	for rows.Next() {
		var e events.EventView
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &startsAt, &participantsCount, &e.EventType); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetAllArchive(ctx context.Context, role string) ([]*events.EventView, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, e.status, (SELECT COUNT(*) FROM event_participants ep WHERE ep.event_id = e.id) AS participants_count, e.starts_at, e.type
		FROM events e
		JOIN event_roles er ON er.event_id = e.id AND (er.event_type = e.type OR er.event_type IS NULL)
		JOIN roles r ON r.id = er.role_id
		WHERE r.name = $1 AND e.ends_at < NOW() AND e.status = 'ACTIVE'
	`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*events.EventView
	for rows.Next() {
		var e events.EventView
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt, &e.EventType); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetEventLocationData(ctx context.Context, id uuid.UUID) (*events.EventView, error) {
	var e events.EventView
	tableName, eventType, err := r.resolveTableName(ctx, id)
	if err != nil {
		return nil, err
	}
	var lat *float64
	var lon *float64
	row := r.pool.QueryRow(ctx, eventLocationSelect(tableName, eventType), id)
	if err := row.Scan(&e.ID, &lat, &lon); err != nil {
		return nil, err
	}
	e.Lat = lat
	e.Long = lon
	e.EventType = eventType
	return &e, nil
}

func (r *eventRepositoryPgx) GetUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userVkId int64) (bool, error) {
	var exists bool
	row := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM event_participants ep
			INNER JOIN users u ON ep.user_id = u.id
			WHERE ep.event_id = $1 AND u.vk_id = $2
		)
	`, eventId, userVkId)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *eventRepositoryPgx) GetEventOrgs(ctx context.Context, eventId uuid.UUID) ([]model.User, error) {
	eventType, err := r.resolveEventType(ctx, eventId)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM event_orgs eo
		JOIN users u ON u.id = eo.user_id
		WHERE eo.event_id = $1 AND (eo.event_type = $2 OR eo.event_type IS NULL)
	`, eventId, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.VkID, &u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetByVkPollId(ctx context.Context, vkPollId int64) (*events.EventView, error) {
	var e events.EventView
	row := r.pool.QueryRow(ctx, `SELECT id, vk_poll_answer_id, 'as_event' FROM event_as_events WHERE vk_vote_id = $1`, vkPollId)
	if err := row.Scan(&e.ID, &e.VkPollAnswerID, &e.EventType); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepositoryPgx) GetEventById(ctx context.Context, id uuid.UUID) (*model.EventAsEvent, error) {
	var e model.EventAsEvent
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, cover, status, starts_at, ends_at, created_at, updated_at,
		       vk_post_id, vk_vote_id, vk_poll_answer_id, lat, lon, address, additional_address
		FROM event_as_events
		WHERE id = $1
	`, id)

	if err := row.Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.Cover,
		&e.Status,
		&e.StartsAt,
		&e.EndsAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.VkPostID,
		&e.VkVoteID,
		&e.VkPollAnswerID,
		&e.Lat,
		&e.Lon,
		&e.Address,
		&e.AdditionalAddress,
	); err != nil {
		return nil, err
	}

	relations, err := r.loadEventRelations(ctx, id, model.EventTypeEvent)
	if err != nil {
		return nil, err
	}
	e.EventRelations = relations

	return &e, nil
}

func (r *eventRepositoryPgx) GetActivityById(ctx context.Context, id uuid.UUID) (*model.EventAsActivity, error) {
	var e model.EventAsActivity
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, cover, status, starts_at, ends_at, created_at, updated_at,
		       train_params, available_activities
		FROM event_as_activities
		WHERE id = $1
	`, id)

	if err := row.Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.Cover,
		&e.Status,
		&e.StartsAt,
		&e.EndsAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.TrainParams,
		&e.AvailableActivities,
	); err != nil {
		return nil, err
	}

	relations, err := r.loadEventRelations(ctx, id, model.EventTypeActivity)
	if err != nil {
		return nil, err
	}
	e.EventRelations = relations

	return &e, nil
}

func (r *eventRepositoryPgx) GetTeamById(ctx context.Context, id uuid.UUID) (*model.EventAsTeam, error) {
	var e model.EventAsTeam
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, cover, status, starts_at, ends_at, created_at, updated_at,
		       teams_constraint, teams_cap_min, teams_cap_max, lat, lon, address, additional_address, vk_post_id
		FROM event_as_teams
		WHERE id = $1
	`, id)

	if err := row.Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.Cover,
		&e.Status,
		&e.StartsAt,
		&e.EndsAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.TeamsConstraint,
		&e.TeamsCapMin,
		&e.TeamsCapMax,
		&e.Lat,
		&e.Lon,
		&e.Address,
		&e.AdditionalAddress,
		&e.VkPostID,
	); err != nil {
		return nil, err
	}

	relations, err := r.loadEventRelations(ctx, id, model.EventTypeTeam)
	if err != nil {
		return nil, err
	}
	e.EventRelations = relations

	return &e, nil
}

func (r *eventRepositoryPgx) GetPollById(ctx context.Context, id uuid.UUID) (*model.EventAsPoll, error) {
	var e model.EventAsPoll
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, cover, status, starts_at, ends_at, created_at, updated_at,
		       ext_link_id, vk_post_id
		FROM event_as_polls
		WHERE id = $1
	`, id)

	if err := row.Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.Cover,
		&e.Status,
		&e.StartsAt,
		&e.EndsAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.ExtLinkID,
		&e.VkPostID,
	); err != nil {
		return nil, err
	}

	relations, err := r.loadEventRelations(ctx, id, model.EventTypePoll)
	if err != nil {
		return nil, err
	}
	e.EventRelations = relations

	return &e, nil
}

func (r *eventRepositoryPgx) GetTestById(ctx context.Context, id uuid.UUID) (*model.EventAsTest, error) {
	var e model.EventAsTest
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, cover, status, starts_at, ends_at, created_at, updated_at,
		       ext_link_id, attempts, event_id, vk_post_id
		FROM event_as_tests
		WHERE id = $1
	`, id)

	if err := row.Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.Cover,
		&e.Status,
		&e.StartsAt,
		&e.EndsAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.ExtLinkID,
		&e.Attempts,
		&e.EventID,
		&e.VkPostID,
	); err != nil {
		return nil, err
	}

	relations, err := r.loadEventRelations(ctx, id, model.EventTypeTest)
	if err != nil {
		return nil, err
	}
	e.EventRelations = relations

	return &e, nil
}

func (r *eventRepositoryPgx) GetLocationById(ctx context.Context, id uuid.UUID) (*events.EventView, error) {
	var e events.EventView
	tableName, eventType, err := r.resolveTableName(ctx, id)
	if err != nil {
		return nil, err
	}
	var lat *float64
	var lon *float64
	row := r.pool.QueryRow(ctx, eventLocationSelect(tableName, eventType), id)
	if err := row.Scan(&e.ID, &lat, &lon); err != nil {
		return nil, err
	}
	e.Lat = lat
	e.Long = lon
	e.EventType = eventType
	return &e, nil
}

func (r *eventRepositoryPgx) GetParticipants(ctx context.Context, id uuid.UUID) ([]model.EventParticipant, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ep.id, ep.user_id, ep.event_id, ep.is_checked, ep.check_timestamp, u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM event_participants ep
		JOIN users u ON u.id = ep.user_id
		WHERE ep.event_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []model.EventParticipant
	for rows.Next() {
		var ep model.EventParticipant
		var checkTimestamp *time.Time
		var u model.User
		if err := rows.Scan(&ep.ID, &ep.UserID, &ep.EventID, &ep.IsChecked, &checkTimestamp, &u.ID, &u.VkID, &u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
			return nil, err
		}
		ep.CheckTimestamp = checkTimestamp
		ep.User = u
		res = append(res, ep)
	}
	return res, nil
}

func (r *eventRepositoryPgx) Create(ctx context.Context, event *events.EventView) (*events.EventView, error) {
	// begin tx
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	eventType := event.EventType
	if eventType == "" {
		eventType = defaultEventType
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO event_as_events (id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, event.ID, event.Name, event.Cover, event.Description, event.Address, event.AdditionalAddress, event.VkPostId, event.VkVoteID, event.VkPollAnswerID, event.Status, event.StartsAt, event.EndsAt, event.Lat, event.Long)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("failed to save as_event: %w", err)
	}

	for _, org := range event.Orgs {
		_, err := tx.Exec(ctx, `INSERT INTO event_orgs (event_id, user_id, event_type) VALUES ($1, $2, $3)`, event.ID, org.ID, eventType)
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to append orgs: %w", err)
		}
	}

	for _, role := range event.AvailableRoles {
		_, err := tx.Exec(ctx, `INSERT INTO event_roles (event_id, role_id, event_type) VALUES ($1, $2, $3)`, event.ID, role.ID, eventType)
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to append roles: %w", err)
		}
	}

	var created events.EventView
	row := tx.QueryRow(ctx, `
		SELECT id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon
		FROM event_as_events WHERE id = $1
	`, event.ID)
	var cover *string
	var description *string
	var address *string
	var additionalAddress *string
	var vkPostId *int64
	var vkVoteId *int64
	var vkPollAns *int64
	var status *string
	var startsAt *time.Time
	var endsAt *time.Time
	var lat *float64
	var lon *float64
	if err := row.Scan(&created.ID, &created.Name, &cover, &description, &address, &additionalAddress, &vkPostId, &vkVoteId, &vkPollAns, &status, &startsAt, &endsAt, &lat, &lon); err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("failed to load created as_event: %w", err)
	}
	created.EventType = eventType
	created.Cover = cover
	created.Description = description
	created.Address = address
	created.AdditionalAddress = additionalAddress
	created.VkPostId = vkPostId
	created.VkVoteID = vkVoteId
	created.VkPollAnswerID = vkPollAns
	created.Status = status
	created.StartsAt = startsAt
	created.EndsAt = endsAt
	created.Lat = lat
	created.Long = lon

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	// load associations
	created.Orgs, _ = r.GetEventOrgs(ctx, created.ID)
	// roles and attachments not strictly necessary here; left nil if not needed

	return &created, nil
}

func (r *eventRepositoryPgx) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*events.EventView, error) {
	// begin tx
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	tableName, eventType, err := r.resolveTableName(ctx, id)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	// check exists
	var exists bool
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)`, tableName), id).Scan(&exists); err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("as_event check failed: %w", err)
	}
	if !exists {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("as_event not found")
	}

	if len(fields) > 0 {
		set := ""
		args := make([]interface{}, 0, len(fields)+1)
		i := 1
		for k, v := range fields {
			if set != "" {
				set += ", "
			}
			set += fmt.Sprintf("%s = $%d", k, i)
			args = append(args, v)
			i++
		}
		args = append(args, id)
		q := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, set, i)
		if _, err := tx.Exec(ctx, q, args...); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to update as_event: %w", err)
		}
	}

	if orgs != nil {
		// replace orgs
		if _, err := tx.Exec(ctx, `DELETE FROM event_orgs WHERE event_id = $1 AND (event_type = $2 OR event_type IS NULL)`, id, eventType); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to clear orgs: %w", err)
		}
		for _, o := range orgs {
			if _, err := tx.Exec(ctx, `INSERT INTO event_orgs (event_id, user_id, event_type) VALUES ($1, $2, $3)`, id, o.ID, eventType); err != nil {
				tx.Rollback(ctx)
				return nil, fmt.Errorf("failed to replace orgs: %w", err)
			}
		}
	}

	if availableRoles != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM event_roles WHERE event_id = $1 AND (event_type = $2 OR event_type IS NULL)`, id, eventType); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to clear roles: %w", err)
		}
		for _, rle := range availableRoles {
			if _, err := tx.Exec(ctx, `INSERT INTO event_roles (event_id, role_id, event_type) VALUES ($1, $2, $3)`, id, rle.ID, eventType); err != nil {
				tx.Rollback(ctx)
				return nil, fmt.Errorf("failed to replace roles: %w", err)
			}
		}
	}

	var updated events.EventView
	row := tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon
		FROM %s WHERE id = $1
	`, tableName), id)
	var cover *string
	var description *string
	var address *string
	var additionalAddress *string
	var vkPostId *int64
	var vkVoteId *int64
	var vkPollAns *int64
	var status *string
	var startsAt *time.Time
	var endsAt *time.Time
	var lat *float64
	var lon *float64
	if err := row.Scan(&updated.ID, &updated.Name, &cover, &description, &address, &additionalAddress, &vkPostId, &vkVoteId, &vkPollAns, &status, &startsAt, &endsAt, &lat, &lon); err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("failed to load updated as_event: %w", err)
	}
	updated.EventType = eventType
	updated.Cover = cover
	updated.Description = description
	updated.Address = address
	updated.AdditionalAddress = additionalAddress
	updated.VkPostId = vkPostId
	updated.VkVoteID = vkVoteId
	updated.VkPollAnswerID = vkPollAns
	updated.Status = status
	updated.StartsAt = startsAt
	updated.EndsAt = endsAt
	updated.Lat = lat
	updated.Long = lon

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	// load associations
	updated.Orgs, _ = r.GetEventOrgs(ctx, updated.ID)

	return &updated, nil
}

func (r *eventRepositoryPgx) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	tableName, _, err := r.resolveTableName(ctx, id)
	if err != nil {
		return err
	}

	cmd, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, tableName), id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	if cmd.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return errors.New("not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
