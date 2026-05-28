package impl

import (
	"context"
	"ne_noy/internal/model"
	"ne_noy/internal/model/events"
	"ne_noy/internal/model/events/as_team"
	"ne_noy/internal/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eventTeamRepositoryPgx struct {
	pool *pgxpool.Pool
}

func NewEventTeamRepository(pool *pgxpool.Pool) repository.EventTeamRepository {
	return &eventTeamRepositoryPgx{pool: pool}
}

func (e *eventTeamRepositoryPgx) GetEventByID(ctx context.Context, eventID uuid.UUID) (*as_team.AsTeam, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT
			id, name, description, cover, status, starts_at, ends_at,
			teams_constraint, teams_cap_min, teams_cap_max,
			lat, lon, address, additional_address, vk_post_id,
			created_at, updated_at
		FROM event_as_teams
		WHERE id = $1
	`, eventID)

	var event as_team.AsTeam
	if err := row.Scan(
		&event.ID, &event.Name, &event.Description, &event.Cover, &event.Status, &event.StartsAt, &event.EndsAt,
		&event.TeamsConstraint, &event.TeamsCapMin, &event.TeamsCapMax,
		&event.Lat, &event.Lon, &event.Address, &event.AdditionalAddress, &event.VkPostID,
		&event.CreatedAt, &event.UpdatedAt,
	); err != nil {
		return nil, err
	}

	attachments, err := e.getEventAttachments(ctx, eventID)
	if err != nil {
		return nil, err
	}
	event.Attachments = attachments

	roles, err := e.getEventRoles(ctx, eventID)
	if err != nil {
		return nil, err
	}
	event.AvailableRoleCodes = roles

	return &event, nil
}

func (e *eventTeamRepositoryPgx) CreateEvent(ctx context.Context, event *as_team.AsTeam) (*as_team.AsTeam, error) {
	eventID := uuid.New()

	// Профиль командного мероприятия хранится отдельно от команд; команды будут ссылаться на этот ID.
	_, err := e.pool.Exec(ctx, `
		INSERT INTO event_as_teams (
			id, name, description, cover, status, starts_at, ends_at,
			teams_constraint, teams_cap_min, teams_cap_max,
			lat, lon, address, additional_address, vk_post_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, eventID, event.Name, event.Description, event.Cover, event.Status, event.StartsAt, event.EndsAt,
		event.TeamsConstraint, event.TeamsCapMin, event.TeamsCapMax, event.Lat, event.Lon, event.Address,
		event.AdditionalAddress, event.VkPostID)
	if err != nil {
		return nil, err
	}

	if len(event.Attachments) > 0 {
		if err = e.replaceEventAttachments(ctx, eventID, event.Attachments); err != nil {
			return nil, err
		}
	}

	if len(event.AvailableRoleCodes) > 0 {
		if err = e.replaceEventRoles(ctx, eventID, event.AvailableRoleCodes); err != nil {
			return nil, err
		}
	}

	return e.GetEventByID(ctx, eventID)
}

func (e *eventTeamRepositoryPgx) UpdateEvent(ctx context.Context, eventID uuid.UUID, update as_team.AsTeam) (*as_team.AsTeam, error) {
	// COALESCE оставляет текущее значение, если сервис не передал поле в DTO обновления.
	commandTag, err := e.pool.Exec(ctx, `
		UPDATE event_as_teams
		SET
			name = COALESCE($2, name),
			description = COALESCE($3, description),
			cover = COALESCE($4, cover),
			status = COALESCE($5, status),
			starts_at = COALESCE($6, starts_at),
			ends_at = COALESCE($7, ends_at),
			teams_constraint = COALESCE($8, teams_constraint),
			teams_cap_min = COALESCE($9, teams_cap_min),
			teams_cap_max = COALESCE($10, teams_cap_max),
			lat = COALESCE($11, lat),
			lon = COALESCE($12, lon),
			address = COALESCE($13, address),
			additional_address = COALESCE($14, additional_address),
			vk_post_id = COALESCE($15, vk_post_id),
			updated_at = now()
		WHERE id = $1
	`, eventID, nullableString(update.Name), update.Description, update.Cover, nullableString(update.Status),
		nullableTime(update.StartsAt), update.EndsAt, nullableInt(update.TeamsConstraint), update.TeamsCapMin,
		update.TeamsCapMax, update.Lat, update.Lon, update.Address, update.AdditionalAddress, update.VkPostID)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}

	if update.Attachments != nil {
		if err := e.replaceEventAttachments(ctx, eventID, update.Attachments); err != nil {
			return nil, err
		}
	}

	if update.AvailableRoleCodes != nil {
		if err := e.replaceEventRoles(ctx, eventID, update.AvailableRoleCodes); err != nil {
			return nil, err
		}
	}

	return e.GetEventByID(ctx, eventID)
}

func (e *eventTeamRepositoryPgx) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Сначала удаляем участников команд, потому что они ссылаются на teams.id.
	if _, err = tx.Exec(ctx, `
		DELETE FROM team_members
		WHERE team_id IN (SELECT id FROM teams WHERE event_id = $1)
	`, eventID); err != nil {
		return err
	}

	// Затем удаляем сами команды, связанные с командным мероприятием.
	if _, err = tx.Exec(ctx, `
		DELETE FROM teams
		WHERE event_id = $1
	`, eventID); err != nil {
		return err
	}

	commandTag, err := tx.Exec(ctx, `
		DELETE FROM event_as_teams
		WHERE id = $1
	`, eventID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}

func (e *eventTeamRepositoryPgx) GetTeamsByEvent(ctx context.Context, eventID uuid.UUID) ([]as_team.Team, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT
			t.id, t.created_at, t.updated_at, t.captain_id, t.event_id, t.team_name,
			c.id, c.created_at, c.vk_id, c.first_name, c.last_name, c.role_id, c.photo_url,
			c.geo_available, c.notification_available,
			COUNT(tm.id) AS total_members
		FROM teams t
		INNER JOIN users c ON c.id = t.captain_id
		LEFT JOIN team_members tm ON tm.team_id = t.id
		WHERE t.event_id = $1
		GROUP BY t.id, c.id
		ORDER BY t.team_name ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]as_team.Team, 0)
	for rows.Next() {
		team, _, scanErr := scanTeamWithCaptain(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		team.Members, scanErr = e.getTeamMembers(ctx, team.ID)
		if scanErr != nil {
			return nil, scanErr
		}
		teams = append(teams, *team)
	}

	return teams, rows.Err()
}

func (e *eventTeamRepositoryPgx) GetTeamByID(ctx context.Context, teamID uuid.UUID) (*as_team.Team, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT
			t.id, t.created_at, t.updated_at, t.captain_id, t.event_id, t.team_name,
			c.id, c.created_at, c.vk_id, c.first_name, c.last_name, c.role_id, c.photo_url,
			c.geo_available, c.notification_available,
			COUNT(tm.id) AS total_members
		FROM teams t
		INNER JOIN users c ON c.id = t.captain_id
		LEFT JOIN team_members tm ON tm.team_id = t.id
		WHERE t.id = $1
		GROUP BY t.id, c.id
	`, teamID)

	team, _, err := scanTeamWithCaptain(row)
	if err != nil {
		return nil, err
	}
	team.Members, err = e.getTeamMembers(ctx, team.ID)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (e *eventTeamRepositoryPgx) CreateTeam(ctx context.Context, eventID, captainID uuid.UUID, name string) (*as_team.Team, error) {
	teamID := uuid.New()

	// Команда создается отдельной записью, капитан хранится в teams.captain_id и не дублируется в team_members.
	_, err := e.pool.Exec(ctx, `
		INSERT INTO teams (id, captain_id, event_id, team_name)
		VALUES ($1, $2, $3, $4)
	`, teamID, captainID, eventID, name)
	if err != nil {
		return nil, err
	}

	return e.GetTeamByID(ctx, teamID)
}

func (e *eventTeamRepositoryPgx) DeleteTeam(ctx context.Context, eventID uuid.UUID) error {
	_, err := e.pool.Exec(ctx, `DELETE FROM teams WHERE id = $1`, eventID)
	return err
}

func (e *eventTeamRepositoryPgx) AddMember(ctx context.Context, teamID, userID uuid.UUID) error {
	team, err := e.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team.CaptainID == userID {
		return nil
	}

	// ON CONFLICT делает повторное вступление идемпотентным и защищает от гонок параллельных запросов.
	_, err = e.pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (team_id, user_id) DO NOTHING
	`, teamID, userID)
	return err
}

func (e *eventTeamRepositoryPgx) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	team, err := e.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}
	if team.CaptainID == userID {
		return pgx.ErrNoRows
	}

	commandTag, err := e.pool.Exec(ctx, `
		DELETE FROM team_members
		WHERE team_id = $1 AND user_id = $2
	`, teamID, userID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (e *eventTeamRepositoryPgx) getTeamMembers(ctx context.Context, teamID uuid.UUID) ([]as_team.TeamMember, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT
			tm.id, tm.created_at, tm.updated_at, tm.team_id, tm.user_id,
			u.id, u.created_at, u.vk_id, u.first_name, u.last_name, u.role_id, u.photo_url,
			u.geo_available, u.notification_available
		FROM team_members tm
		INNER JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY tm.created_at ASC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]as_team.TeamMember, 0)
	for rows.Next() {
		var member as_team.TeamMember
		if err = rows.Scan(
			&member.ID, &member.CreatedAt, &member.UpdatedAt, &member.TeamID, &member.UserID,
			&member.User.ID, &member.User.CreatedAt, &member.User.VkID, &member.User.FirstName, &member.User.LastName,
			&member.User.RoleID, &member.User.PhotoURL, &member.User.GeoAvailable, &member.User.NotificationAvailable,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

func (e *eventTeamRepositoryPgx) UpdateCaptain(ctx context.Context, teamID, newCaptainID uuid.UUID) error {
	team, err := e.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	oldCaptainID := team.CaptainID
	if oldCaptainID == newCaptainID {
		return nil
	}

	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Новый капитан должен быть участником команды — убираем его из team_members.
	if _, err = tx.Exec(ctx, `
		DELETE FROM team_members WHERE team_id = $1 AND user_id = $2
	`, teamID, newCaptainID); err != nil {
		return err
	}

	// Старый капитан становится рядовым участником.
	if _, err = tx.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)
		ON CONFLICT (team_id, user_id) DO NOTHING
	`, teamID, oldCaptainID); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `
		UPDATE teams SET captain_id = $2, updated_at = now() WHERE id = $1
	`, teamID, newCaptainID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (e *eventTeamRepositoryPgx) SetEventOrganizers(ctx context.Context, eventID uuid.UUID, userIDs []uuid.UUID) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM event_orgs WHERE event_id = $1 AND event_type = 'team'`, eventID); err != nil {
		return err
	}

	for _, userID := range userIDs {
		if _, err = tx.Exec(ctx, `
			INSERT INTO event_orgs (event_id, event_type, user_id) VALUES ($1, 'team', $2)
			ON CONFLICT DO NOTHING
		`, eventID, userID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (e *eventTeamRepositoryPgx) GetEventOrganizers(ctx context.Context, eventID uuid.UUID) ([]model.User, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT u.id, u.vk_id, u.first_name, u.last_name, u.photo_url
		FROM users u
		JOIN event_orgs eo ON eo.user_id = u.id
		WHERE eo.event_id = $1 AND eo.event_type = 'team'
		ORDER BY u.first_name ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.VkID, &u.FirstName, &u.LastName, &u.PhotoURL); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (e *eventTeamRepositoryPgx) getEventAttachments(ctx context.Context, eventID uuid.UUID) ([]events.EventAttachment, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT ea.attachment_id, a.url, a.filename
		FROM event_attachments ea
		JOIN attachments a ON a.id = ea.attachment_id
		WHERE ea.event_id = $1 AND ea.event_type = $2::event_type_enum
		ORDER BY ea.created_at ASC
	`, eventID, events.EventAsTeam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]events.EventAttachment, 0)
	for rows.Next() {
		var ea events.EventAttachment
		ea.EventID = eventID
		ea.EventType = events.EventAsTeam
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

func (e *eventTeamRepositoryPgx) replaceEventAttachments(ctx context.Context, eventID uuid.UUID, attachments []events.EventAttachment) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_attachments WHERE event_id = $1 AND event_type = $2::event_type_enum
	`, eventID, events.EventAsTeam)
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
		`, eventID, events.EventAsTeam, attachmentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *eventTeamRepositoryPgx) replaceEventRoles(ctx context.Context, eventID uuid.UUID, roleCodes []string) error {
	_, err := e.pool.Exec(ctx, `
		DELETE FROM event_roles WHERE event_id = $1 AND event_type = 'team'::event_type_enum
	`, eventID)
	if err != nil {
		return err
	}
	if len(roleCodes) == 0 {
		return nil
	}
	_, err = e.pool.Exec(ctx, `
		INSERT INTO event_roles (event_id, event_type, role_id)
		SELECT $1, 'team'::event_type_enum, r.id
		FROM roles r
		WHERE r.name = ANY($2)
		ON CONFLICT DO NOTHING
	`, eventID, roleCodes)
	return err
}

func (e *eventTeamRepositoryPgx) getEventRoles(ctx context.Context, eventID uuid.UUID) ([]string, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT r.name
		FROM event_roles er
		JOIN roles r ON r.id = er.role_id
		WHERE er.event_id = $1 AND er.event_type = 'team'::event_type_enum
		ORDER BY r.name ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	codes := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

func scanTeamWithCaptain(row pgx.Row) (*as_team.Team, int64, error) {
	var (
		team             as_team.Team
		captainCreatedAt time.Time
		totalMembers     int64
	)

	if err := row.Scan(
		&team.ID, &team.CreatedAt, &team.UpdatedAt, &team.CaptainID, &team.EventID, &team.TeamName,
		&team.Captain.ID, &captainCreatedAt, &team.Captain.VkID, &team.Captain.FirstName, &team.Captain.LastName,
		&team.Captain.RoleID, &team.Captain.PhotoURL, &team.Captain.GeoAvailable, &team.Captain.NotificationAvailable,
		&totalMembers,
	); err != nil {
		return nil, 0, err
	}

	team.Captain.CreatedAt = captainCreatedAt
	return &team, totalMembers, nil
}
