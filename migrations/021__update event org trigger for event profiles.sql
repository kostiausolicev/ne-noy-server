-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_any_org_exist()
    RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM event_orgs eo
        WHERE eo.event_id = OLD.event_id
          AND eo.event_type = OLD.event_type
    ) THEN
        RETURN NULL;
    END IF;

    CASE OLD.event_type::text
        WHEN 'event' THEN
            DELETE FROM event_as_events WHERE id = OLD.event_id;
        WHEN 'test' THEN
            DELETE FROM event_as_tests WHERE id = OLD.event_id;
        WHEN 'team' THEN
            DELETE FROM event_as_teams WHERE id = OLD.event_id;
        WHEN 'poll' THEN
            DELETE FROM event_as_polls WHERE id = OLD.event_id;
        WHEN 'activity' THEN
            DELETE FROM event_as_activities WHERE id = OLD.event_id;
        ELSE
            RAISE EXCEPTION 'Unsupported event type: %', OLD.event_type;
    END CASE;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_any_org_exist()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM events e
    WHERE e.id = OLD.event_id
      AND NOT EXISTS (
        SELECT 1
        FROM event_orgs eo
        WHERE eo.event_id = e.id
    );

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
