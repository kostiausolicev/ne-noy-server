-- +goose Up

-- Триггер check_any_org_exist удалял мероприятие при отсутствии организаторов.
-- Это ломает паттерн replace (DELETE all + INSERT new): после DELETE всех строк
-- event_orgs триггер видит 0 организаторов и удаляет само мероприятие до того,
-- как успевают вставиться новые организаторы.
-- Проверку "у мероприятия должен быть хотя бы один организатор" выполняет приложение.
DROP TRIGGER IF EXISTS trg_check_any_org_exist ON event_orgs;
DROP FUNCTION IF EXISTS check_any_org_exist();

-- +goose Down
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

CREATE TRIGGER trg_check_any_org_exist
    AFTER DELETE ON event_orgs
    FOR EACH ROW
EXECUTE FUNCTION check_any_org_exist();
