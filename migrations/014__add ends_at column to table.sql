-- +goose Up
ALTER TABLE events ADD COLUMN ends_at TIMESTAMP DEFAULT NOW();

-- +goose Down
ALTER TABLE events DROP COLUMN ends_at;