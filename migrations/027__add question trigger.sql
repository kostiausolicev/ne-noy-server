-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION validateSortOrder() RETURNS TRIGGER AS
$$
BEGIN
    IF NEW.q_order <= 0 THEN
        RAISE EXCEPTION 'q_order must be greater than 0';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER validate_q_order_before_insert
    BEFORE INSERT
    ON questions
    FOR EACH ROW
EXECUTE FUNCTION validateSortOrder();

CREATE TRIGGER validate_q_order_before_update
    BEFORE UPDATE OF q_order
    ON questions
    FOR EACH ROW
EXECUTE FUNCTION validateSortOrder();

-- +goose Down
DROP TRIGGER validate_q_order_before_update ON questions;
DROP TRIGGER validate_q_order_before_insert ON questions;
DROP FUNCTION validateSortOrder();
