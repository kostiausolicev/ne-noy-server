package impl

import (
	"context"
	"database/sql"
	"fmt"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_event"
	"ne_noy/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventEventRepositoryPgx struct {
	pool *pgxpool.Pool
}

func (e *eventEventRepositoryPgx) GetLocationById(ctx context.Context, id uuid.UUID) (lat, long *float64, err error) {
	row := e.pool.QueryRow(ctx, `
		SELECT lat, lon
		FROM event_as_events
		WHERE id = $1
	`, id)
	err = row.Scan(&lat, &long)
	return lat, long, err
}

func (e *eventEventRepositoryPgx) ExistUserParticipationInEvent(ctx context.Context, eventId uuid.UUID, userId uuid.UUID) (bool, error) {
	var exists bool
	err := e.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM event_participants
			WHERE event_id = $1 AND user_id = $2
		)
	`, eventId, userId).Scan(&exists)
	return exists, err
}

func (e *eventEventRepositoryPgx) GetEventOrgs(ctx context.Context, id uuid.UUID, limit int) ([]model.User, error) {
	query := `
		SELECT
			u.id, u.created_at, u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			u.geo_available, u.notification_available
		FROM event_orgs eo
		INNER JOIN users u ON u.id = eo.user_id
		WHERE eo.event_id = $1 AND eo.event_type = $2::event_type_enum
		ORDER BY u.created_at ASC
	`
	args := []interface{}{id, events.EventAsEvent}
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
		user, err := scanEventUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (e *eventEventRepositoryPgx) GetByVkPollId(ctx context.Context, pollId int64) (*as_event.AsEvent, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT
			id, created_at, name, description, cover, status, starts_at, ends_at,
			vk_post_id, vk_vote_id, vk_poll_answer_id, lat, lon, address, additional_address
		FROM event_as_events
		WHERE vk_vote_id = $1
	`, pollId)
	event, err := scanAsEvent(row)
	if err != nil {
		return nil, err
	}
	if err = e.fillEventRelations(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (e *eventEventRepositoryPgx) GetEventById(ctx context.Context, id uuid.UUID) (*as_event.AsEvent, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT
			id, created_at, name, description, cover, status, starts_at, ends_at,
			vk_post_id, vk_vote_id, vk_poll_answer_id, lat, lon, address, additional_address
		FROM event_as_events
		WHERE id = $1
	`, id)
	event, err := scanAsEvent(row)
	if err != nil {
		return nil, err
	}
	if err = e.fillEventRelations(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (e *eventEventRepositoryPgx) GetParticipants(ctx context.Context, id uuid.UUID, limit int) ([]as_event.EventParticipants, error) {
	query := `
		SELECT
			ep.id, ep.created_at, ep.event_id, ep.user_id, ep.prepare_type, ep.is_checked,
			ep.check_timestamp, ep.check_lat, ep.check_lon, ep.check_type, ep.check_author,
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

	participants := make([]as_event.EventParticipants, 0)
	for rows.Next() {
		var participant as_event.EventParticipants
		var checkType sql.NullString
		if err = rows.Scan(
			&participant.ID, &participant.CreatedAt, &participant.EventID, &participant.UserID, &participant.PrepareType,
			&participant.IsChecked, &participant.CheckTimestamp, &participant.CheckLat, &participant.CheckLong,
			&checkType, &participant.CheckAuthor,
			&participant.User.ID, &participant.User.CreatedAt, &participant.User.VkID, &participant.User.FirstName,
			&participant.User.LastName, &participant.User.RoleID, &participant.User.PhotoURL,
			&participant.User.GeoAvailable, &participant.User.NotificationAvailable,
		); err != nil {
			return nil, err
		}
		participant.CheckType = checkType.String
		participants = append(participants, participant)
	}
	return participants, rows.Err()
}

func (e *eventEventRepositoryPgx) CreateEvent(ctx context.Context, event *as_event.AsEvent) (*as_event.AsEvent, error) {
	eventID := uuid.New()
	_, err := e.pool.Exec(ctx, `
		INSERT INTO event_as_events (
			id, name, description, cover, status, starts_at, ends_at,
			vk_post_id, vk_vote_id, vk_poll_answer_id, lat, lon, address, additional_address
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, eventID, event.Name, event.Description, event.Cover, event.Status, event.StartsAt, event.EndsAt,
		event.VkPostID, event.VkVoteID, event.VkPollAnswerID, event.Lat, event.Lon, event.Address, event.AdditionalAddress)
	if err != nil {
		return nil, err
	}
	if err = e.replaceEventOrgs(ctx, eventID, event.Orgs); err != nil {
		return nil, err
	}
	if len(event.AvailableRoleCodes) > 0 {
		if err = e.replaceEventRoles(ctx, eventID, event.AvailableRoleCodes); err != nil {
			return nil, err
		}
	}
	if len(event.Attachments) > 0 {
		if err = e.replaceEventAttachments(ctx, eventID, event.Attachments); err != nil {
			return nil, err
		}
	}
	return e.GetEventById(ctx, eventID)
}

func (e *eventEventRepositoryPgx) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, roleCodes []string, attachments []events.EventAttachment) (*as_event.AsEvent, error) {
	if len(fields) > 0 {
		setParts := make([]string, 0, len(fields))
		args := make([]interface{}, 0, len(fields)+1)
		args = append(args, id)
		i := 2
		for field, value := range fields {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, i))
			args = append(args, value)
			i++
		}
		query := fmt.Sprintf("UPDATE event_as_events SET %s WHERE id = $1", strings.Join(setParts, ", "))
		tag, err := e.pool.Exec(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, pgx.ErrNoRows
		}
	}
	if orgs != nil {
		if err := e.replaceEventOrgs(ctx, id, orgs); err != nil {
			return nil, err
		}
	}
	if roleCodes != nil {
		if err := e.replaceEventRoles(ctx, id, roleCodes); err != nil {
			return nil, err
		}
	}
	if attachments != nil {
		if err := e.replaceEventAttachments(ctx, id, attachments); err != nil {
			return nil, err
		}
	}
	return e.GetEventById(ctx, id)
}

func NewEventEventRepository(db *pgxpool.Pool) repository.EventEventRepository {
	return &eventEventRepositoryPgx{pool: db}
}

func scanAsEvent(row pgx.Row) (*as_event.AsEvent, error) {
	var event as_event.AsEvent
	err := row.Scan(
		&event.ID, &event.CreatedAt, &event.Name, &event.Description, &event.Cover, &event.Status, &event.StartsAt, &event.EndsAt,
		&event.VkPostID, &event.VkVoteID, &event.VkPollAnswerID, &event.Lat, &event.Lon, &event.Address, &event.AdditionalAddress,
	)
	if err != nil {
		return nil, err
	}
	event.StartsAt = event.StartsAt.UTC()
	if event.EndsAt != nil {
		endsAt := event.EndsAt.UTC()
		event.EndsAt = &endsAt
	}
	return &event, nil
}

func (e *eventEventRepositoryPgx) fillEventRelations(ctx context.Context, event *as_event.AsEvent) error {
	var err error
	event.Orgs, err = e.GetEventOrgs(ctx, event.ID, 0)
	if err != nil {
		return err
	}
	event.EventParticipants, err = e.GetParticipants(ctx, event.ID, 3)
	if err != nil {
		return err
	}
	event.Attachments, err = e.getEventAttachments(ctx, event.ID)
	if err != nil {
		return err
	}
	err = e.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM event_participants
		WHERE event_id = $1
	`, event.ID).Scan(&event.ParticipantsCount)
	return err
}

func (e *eventEventRepositoryPgx) getEventAttachments(ctx context.Context, eventID uuid.UUID) ([]events.EventAttachment, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT ea.attachment_id, a.url, a.filename
		FROM event_attachments ea
		JOIN attachments a ON a.id = ea.attachment_id
		WHERE ea.event_id = $1 AND ea.event_type = $2::event_type_enum
		ORDER BY ea.created_at ASC
	`, eventID, events.EventAsEvent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]events.EventAttachment, 0)
	for rows.Next() {
		var ea events.EventAttachment
		ea.EventID = eventID
		ea.EventType = events.EventAsEvent
		ea.Attachment = &model.Attachment{}
		if err := rows.Scan(&ea.AttachmentID, &ea.Attachment.Url, &ea.Attachment.Filename); err != nil {
			return nil, err
		}
		if ea.AttachmentID != nil {
			ea.Attachment.ID = *ea.AttachmentID
		}
		result = append(result, ea)
	}
	return result, rows.Err()
}

func (e *eventEventRepositoryPgx) replaceEventAttachments(ctx context.Context, eventID uuid.UUID, attachments []events.EventAttachment) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_attachments WHERE event_id = $1 AND event_type = $2::event_type_enum
	`, eventID, events.EventAsEvent)
	if err != nil {
		return err
	}
	for _, a := range attachments {
		if a.Attachment == nil {
			continue
		}
		var attachmentID int64
		if a.AttachmentID != nil && *a.AttachmentID != 0 {
			_, err = e.pool.Exec(ctx, `
				INSERT INTO attachments (id, url, filename)
				VALUES ($1, $2, $3)
				ON CONFLICT (id) DO NOTHING
			`, *a.AttachmentID, a.Attachment.Url, a.Attachment.Filename)
			if err != nil {
				return err
			}
			attachmentID = *a.AttachmentID
		} else {
			err = e.pool.QueryRow(ctx, `
				INSERT INTO attachments (url, filename)
				VALUES ($1, $2)
				RETURNING id
			`, a.Attachment.Url, a.Attachment.Filename).Scan(&attachmentID)
			if err != nil {
				return err
			}
		}
		_, err = e.pool.Exec(ctx, `
			INSERT INTO event_attachments (event_id, event_type, attachment_id)
			VALUES ($1, $2::event_type_enum, $3)
		`, eventID, events.EventAsEvent, attachmentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *eventEventRepositoryPgx) replaceEventRoles(ctx context.Context, eventID uuid.UUID, roleCodes []string) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_roles WHERE event_id = $1 AND event_type = 'event'::event_type_enum
	`, eventID)
	if err != nil {
		return err
	}
	if len(roleCodes) == 0 {
		return nil
	}
	_, err = e.pool.Exec(ctx, `
		INSERT INTO event_roles (event_id, event_type, role_id)
		SELECT $1, 'event'::event_type_enum, r.id
		FROM roles r
		WHERE r.name = ANY($2)
		ON CONFLICT DO NOTHING
	`, eventID, roleCodes)
	return err
}

func (e *eventEventRepositoryPgx) replaceEventOrgs(ctx context.Context, eventID uuid.UUID, orgs []model.User) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_orgs
		WHERE event_id = $1 AND event_type = $2::event_type_enum
	`, eventID, events.EventAsEvent)
	if err != nil {
		return err
	}
	for _, org := range orgs {
		if org.ID == uuid.Nil {
			continue
		}
		_, err = e.pool.Exec(ctx, `
			INSERT INTO event_orgs (event_id, event_type, user_id)
			VALUES ($1, $2::event_type_enum, $3)
			ON CONFLICT DO NOTHING
		`, eventID, events.EventAsEvent, org.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func scanEventUser(row pgx.Row) (model.User, error) {
	var user model.User
	err := row.Scan(
		&user.ID, &user.CreatedAt, &user.VkID, &user.FirstName, &user.LastName,
		&user.RoleID, &user.PhotoURL, &user.GeoAvailable, &user.NotificationAvailable,
	)
	return user, err
}
