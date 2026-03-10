-- +goose Up
ALTER TABLE user_watches_onbordings
    RENAME TO user_watches_onboardings;

ALTER TABLE user_watches_onboardings
    RENAME COLUMN onbording_id TO onboarding_id;

-- +goose Down
ALTER TABLE user_watches_onboardings
    RENAME TO user_watches_onbordings;

ALTER TABLE user_watches_onbordings
    RENAME COLUMN onboarding_id TO onbording_id;
