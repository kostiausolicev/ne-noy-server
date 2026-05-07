-- +goose Up
ALTER TABLE team_members
    ADD CONSTRAINT uq_team_members_team_user UNIQUE (team_id, user_id);

-- +goose Down
ALTER TABLE team_members
    DROP CONSTRAINT uq_team_members_team_user;
