-- +goose Up
ALTER TABLE onbordings RENAME TO onboardings;

ALTER TABLE onboardings
    ADD COLUMN IF NOT EXISTS path VARCHAR(255);

-- +goose Down
ALTER TABLE onboardings
    DROP COLUMN IF EXISTS path;

ALTER TABLE onboardings RENAME TO onbordings;