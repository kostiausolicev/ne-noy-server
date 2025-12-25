-- +goose Up

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

CREATE TRIGGER trg_check_any_org_exist
    AFTER DELETE ON event_orgs
    FOR EACH ROW
EXECUTE FUNCTION check_any_org_exist();

-- +goose Down
DROP TRIGGER IF EXISTS trg_check_any_org_exist ON event_orgs;
DROP FUNCTION IF EXISTS check_any_org_exist();
