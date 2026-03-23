-- +goose Up
ALTER TABLE event_participants ADD CONSTRAINT participant_check_date_constraint CHECK (
    check_timestamp < (SELECT e1.ends_at FROM events e1 WHERE e1.id = event_id)
);

-- +goose Down
ALTER TABLE event_participants DROP CONSTRAINT participant_check_date_constraint;