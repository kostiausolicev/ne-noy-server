-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_check_date_trigger() RETURNS TRIGGER AS
$$
DECLARE
    eventEndsAt TIMESTAMP;
BEGIN
    SELECT e.ends_at
    INTO eventEndsAt
    FROM events e
    WHERE e.id = NEW.event_id;

    IF NEW.check_timestamp IS NOT NULL AND NEW.check_timestamp > eventEndsAt THEN
        RAISE NOTICE 'Время отметки на мероприятии не может быть позже даты завершения мероприятия';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER set_check_timestamp_check_trigger
    BEFORE UPDATE
    ON event_participants
    FOR EACH ROW
EXECUTE FUNCTION check_check_date_trigger();

-- +goose Down
DROP TRIGGER set_check_timestamp_check_trigger ON event_participants;
DROP FUNCTION check_check_date_trigger();