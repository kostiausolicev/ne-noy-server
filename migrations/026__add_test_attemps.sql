-- +goose Up
CREATE TABLE user_attempts
(
    id      uuid PRIMARY KEY,
    userid  uuid REFERENCES users (id) ON DELETE CASCADE,
    testid  uuid REFERENCES event_as_tests (id) ON DELETE CASCADE,
    started TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

ALTER TABLE user_answers
    ADD COLUMN attempt uuid REFERENCES user_attempts (id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE user_answers
    DROP COLUMN attempt;
DROP TABLE user_attempts;