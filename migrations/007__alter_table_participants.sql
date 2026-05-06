-- +goose Up
ALTER TABLE event_participants ADD COLUMN IF NOT EXISTS prepare_type VARCHAR(5) NOT NULL DEFAULT 'app';

-- +goose Down
ALTER TABLE event_participants DROP COLUMN IF EXISTS prepare_type;
