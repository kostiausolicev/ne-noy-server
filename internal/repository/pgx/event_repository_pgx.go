package pgx

import (
	"context"
	"errors"
	"fmt"
	"ne_noy/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ne_noy/internal/model"
)

type eventRepositoryPgx struct {
	pool *pgxpool.Pool
}

func NewEventRepositoryPgx(pool *pgxpool.Pool) repository.EventRepository {
	return &eventRepositoryPgx{pool: pool}
}

func (r *eventRepositoryPgx) GetAll(ctx context.Context) ([]*model.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, status, (SELECT COUNT(*) FROM event_participants ep WHERE ep.event_id = events.id) AS participants_count, starts_at
		FROM events
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*model.Event
	for rows.Next() {
		var e model.Event
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetAllByOrg(ctx context.Context, orgId uuid.UUID) ([]*model.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, e.status, (SELECT COUNT(*) FROM event_participants ep WHERE ep.event_id = e.id) AS participants_count, e.starts_at
		FROM events e
		LEFT JOIN event_orgs eo ON eo.event_id = e.id
		WHERE eo.user_id = $1
	`, orgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*model.Event
	for rows.Next() {
		var e model.Event
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetAllByRole(ctx context.Context, role string) ([]*model.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, e.status, (SELECT COUNT(*) FROM event_participants ep WHERE ep.event_id = e.id) AS participants_count, e.starts_at
		FROM events e
		JOIN event_roles er ON er.event_id = e.id
		JOIN roles r ON r.id = er.role_id
		WHERE r.name = $1 AND e.ends_at > NOW() AND e.status = 'ACTIVE'
	`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*model.Event
	for rows.Next() {
		var e model.Event
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetAllArchive(ctx context.Context, role string) ([]*model.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.name, e.status, (SELECT COUNT(*) FROM event_participants ep WHERE ep.event_id = e.id) AS participants_count, e.starts_at
		FROM events e
		JOIN event_roles er ON er.event_id = e.id
		JOIN roles r ON r.id = er.role_id
		WHERE r.name = $1 AND e.ends_at < NOW() AND e.status = 'ACTIVE'
	`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*model.Event
	for rows.Next() {
		var e model.Event
		var participantsCount int
		var startsAt *time.Time
		if err := rows.Scan(&e.ID, &e.Name, &e.Status, &participantsCount, &startsAt); err != nil {
			return nil, err
		}
		e.ParticipantsCount = participantsCount
		e.StartsAt = startsAt
		res = append(res, &e)
	}
	return res, nil
}

func (r *eventRepositoryPgx) GetEventLocationData(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var e model.Event
	var lat *float64
	var lon *float64
	row := r.pool.QueryRow(ctx, `SELECT id, lat, lon FROM events WHERE id = $1`, id)
	if err := row.Scan(&e.ID, &lat, &lon); err != nil {
		return nil, err
	}
	e.Lat = lat
	e.Long = lon
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
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM event_orgs eo
		JOIN users u ON u.id = eo.user_id
		WHERE eo.event_id = $1
	`, eventId)
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

func (r *eventRepositoryPgx) GetByVkPollId(ctx context.Context, vkPollId int64) (*model.Event, error) {
	var e model.Event
	row := r.pool.QueryRow(ctx, `SELECT id, vk_poll_answer_id FROM events WHERE vk_vote_id = $1`, vkPollId)
	if err := row.Scan(&e.ID, &e.VkPollAnswerID); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepositoryPgx) GetById(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var e model.Event
	// load main fields
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon
		FROM events WHERE id = $1
	`, id)

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

	if err := row.Scan(&e.ID, &e.Name, &cover, &description, &address, &additionalAddress, &vkPostId, &vkVoteId, &vkPollAns, &status, &startsAt, &endsAt, &lat, &lon); err != nil {
		return nil, err
	}
	e.Cover = cover
	e.Description = description
	e.Address = address
	e.AdditionalAddress = additionalAddress
	e.VkPostId = vkPostId
	e.VkVoteID = vkVoteId
	e.VkPollAnswerID = vkPollAns
	e.Status = status
	e.StartsAt = startsAt
	e.EndsAt = endsAt
	e.Lat = lat
	e.Long = lon

	// load orgs
	orgs, err := r.GetEventOrgs(ctx, id)
	if err != nil {
		return nil, err
	}
	e.Orgs = orgs

	// load participants (limit 3 as original)
	pRows, err := r.pool.Query(ctx, `
		SELECT ep.id, ep.user_id, ep.event_id, ep.is_checked, ep.check_timestamp, u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM event_participants ep
		JOIN users u ON u.id = ep.user_id
		WHERE ep.event_id = $1
		ORDER BY ep.created_at DESC
		LIMIT 3
	`, id)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	var participants []model.EventParticipant
	for pRows.Next() {
		var ep model.EventParticipant
		var checkTimestamp *time.Time
		var u model.User
		if err := pRows.Scan(&ep.ID, &ep.UserID, &ep.EventID, &ep.IsChecked, &checkTimestamp, &u.ID, &u.VkID, &u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
			return nil, err
		}
		ep.CheckTimestamp = checkTimestamp
		ep.User = u
		participants = append(participants, ep)
	}
	e.EventParticipants = participants

	// load attachments
	attRows, err := r.pool.Query(ctx, `
		SELECT ea.id, a.id, a.url, a.filename FROM event_attachments ea
		JOIN attachments a ON a.id = ea.attachment_id
		WHERE ea.event_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer attRows.Close()

	var atts []model.EventAttachment
	for attRows.Next() {
		var a model.EventAttachment
		var attachID int64
		var url string
		var filename string
		if err := attRows.Scan(&a.ID, &attachID, &url, &filename); err != nil {
			return nil, err
		}
		// fill inner Attachment
		a.AttachmentID = uuid.UUID{}
		a.Attachment = model.Attachment{ID: attachID, Url: url, Filename: filename}
		atts = append(atts, a)
	}
	e.Attachments = atts

	return &e, nil
}

func (r *eventRepositoryPgx) GetLocationById(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var e model.Event
	var lat *float64
	var lon *float64
	row := r.pool.QueryRow(ctx, `SELECT id, lat, lon FROM events WHERE id = $1`, id)
	if err := row.Scan(&e.ID, &lat, &lon); err != nil {
		return nil, err
	}
	e.Lat = lat
	e.Long = lon
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

func (r *eventRepositoryPgx) Create(ctx context.Context, event *model.Event) (*model.Event, error) {
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

	_, err = tx.Exec(ctx, `
		INSERT INTO events (id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, event.ID, event.Name, event.Cover, event.Description, event.Address, event.AdditionalAddress, event.VkPostId, event.VkVoteID, event.VkPollAnswerID, event.Status, event.StartsAt, event.EndsAt, event.Lat, event.Long)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	for _, org := range event.Orgs {
		_, err := tx.Exec(ctx, `INSERT INTO event_orgs (event_id, user_id) VALUES ($1, $2)`, event.ID, org.ID)
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to append orgs: %w", err)
		}
	}

	for _, role := range event.AvailableRoles {
		_, err := tx.Exec(ctx, `INSERT INTO event_roles (event_id, role_id) VALUES ($1, $2)`, event.ID, role.ID)
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to append roles: %w", err)
		}
	}

	var created model.Event
	row := tx.QueryRow(ctx, `
		SELECT id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon
		FROM events WHERE id = $1
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
		return nil, fmt.Errorf("failed to load created event: %w", err)
	}
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

func (r *eventRepositoryPgx) Update(ctx context.Context, id uuid.UUID, fields map[string]interface{}, orgs []model.User, availableRoles []model.Role) (*model.Event, error) {
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

	// check exists
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`, id).Scan(&exists); err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("event check failed: %w", err)
	}
	if !exists {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("event not found")
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
		q := fmt.Sprintf("UPDATE events SET %s WHERE id = $%d", set, i)
		if _, err := tx.Exec(ctx, q, args...); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to update event: %w", err)
		}
	}

	if orgs != nil {
		// replace orgs
		if _, err := tx.Exec(ctx, `DELETE FROM event_orgs WHERE event_id = $1`, id); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to clear orgs: %w", err)
		}
		for _, o := range orgs {
			if _, err := tx.Exec(ctx, `INSERT INTO event_orgs (event_id, user_id) VALUES ($1, $2)`, id, o.ID); err != nil {
				tx.Rollback(ctx)
				return nil, fmt.Errorf("failed to replace orgs: %w", err)
			}
		}
	}

	if availableRoles != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM event_roles WHERE event_id = $1`, id); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to clear roles: %w", err)
		}
		for _, rle := range availableRoles {
			if _, err := tx.Exec(ctx, `INSERT INTO event_roles (event_id, role_id) VALUES ($1, $2)`, id, rle.ID); err != nil {
				tx.Rollback(ctx)
				return nil, fmt.Errorf("failed to replace roles: %w", err)
			}
		}
	}

	var updated model.Event
	row := tx.QueryRow(ctx, `
		SELECT id, name, cover, description, address, additional_address, vk_post_id, vk_vote_id, vk_poll_answer_id, status, starts_at, ends_at, lat, lon
		FROM events WHERE id = $1
	`, id)
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
		return nil, fmt.Errorf("failed to load updated event: %w", err)
	}
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

	cmd, err := tx.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
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
