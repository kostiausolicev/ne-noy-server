-- +goose Up
ALTER TABLE user_watches_onbordings
    RENAME TO user_watches_onboardings;

-- +goose Down
ALTER TABLE user_watches_onboardings
    RENAME TO user_watches_onbordings;
