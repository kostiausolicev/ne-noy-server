package impl

import (
	"context"
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

	return &event, nil
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
