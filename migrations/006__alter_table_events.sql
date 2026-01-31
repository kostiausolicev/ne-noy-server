-- +goose Up
ALTER TABLE events ADD COLUMN additional_address VARCHAR(255);

-- +goose Down
ALTER TABLE events DROP COLUMN IF EXISTS additional_address;
