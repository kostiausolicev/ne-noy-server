-- +goose Up
CREATE TYPE event_type_enum AS ENUM (
    'as_event',
    'as_test',
    'as_team'
    );

ALTER TABLE event_orgs
    DROP CONSTRAINT fk_event_org_event;
ALTER TABLE event_orgs
    ADD COLUMN event_type event_type_enum NOT NULL DEFAULT 'as_event';

ALTER TABLE event_roles
    DROP CONSTRAINT fk_event_role_event;
ALTER TABLE event_roles
    ADD COLUMN event_type event_type_enum NOT NULL DEFAULT 'as_event';

ALTER TABLE event_attachments
    DROP CONSTRAINT fk_event_attachment_event;
ALTER TABLE event_attachments
    ADD COLUMN event_type event_type_enum NOT NULL DEFAULT 'as_event';

-- +goose Down
ALTER TABLE event_orgs
    DROP COLUMN event_type;
ALTER TABLE event_orgs
    ADD CONSTRAINT fk_event_org_event FOREIGN KEY (event_id) REFERENCES users (id);

ALTER TABLE event_roles
    DROP COLUMN event_type;
ALTER TABLE event_roles
    ADD CONSTRAINT fk_event_role_event FOREIGN KEY (event_id) REFERENCES users (id);

ALTER TABLE event_attachments
    DROP COLUMN event_type;
ALTER TABLE event_attachments
    ADD CONSTRAINT fk_event_attachment_event FOREIGN KEY (event_id) REFERENCES users (id);

DROP TYPE event_type_enum;