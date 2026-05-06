-- +goose Up
CREATE TYPE event_type_enum AS ENUM (
    'event',
    'test',
    'team',
    'poll',
    'activity'
    );

UPDATE event_orgs SET event_type = 'event' WHERE event_type IS NULL;
UPDATE event_roles SET event_type = 'event' WHERE event_type IS NULL;
UPDATE event_attachments SET event_type = 'event' WHERE event_type IS NULL;

ALTER TABLE event_orgs
    ALTER COLUMN event_type TYPE event_type_enum USING event_type::event_type_enum,
    ALTER COLUMN event_type SET DEFAULT 'event',
    ALTER COLUMN event_type SET NOT NULL;

ALTER TABLE event_roles
    ALTER COLUMN event_type TYPE event_type_enum USING event_type::event_type_enum,
    ALTER COLUMN event_type SET DEFAULT 'event',
    ALTER COLUMN event_type SET NOT NULL;

ALTER TABLE event_attachments
    ALTER COLUMN event_type TYPE event_type_enum USING event_type::event_type_enum,
    ALTER COLUMN event_type SET DEFAULT 'event',
    ALTER COLUMN event_type SET NOT NULL;

ALTER TABLE event_orgs
    DROP CONSTRAINT event_orgs_pkey,
    ADD PRIMARY KEY (event_id, event_type, user_id);

ALTER TABLE event_roles
    DROP CONSTRAINT event_roles_pkey,
    ADD PRIMARY KEY (event_id, event_type, role_id);

-- +goose Down
ALTER TABLE event_orgs
    DROP CONSTRAINT event_orgs_pkey,
    ADD PRIMARY KEY (event_id, user_id);

ALTER TABLE event_roles
    DROP CONSTRAINT event_roles_pkey,
    ADD PRIMARY KEY (event_id, role_id);

ALTER TABLE event_orgs
    ALTER COLUMN event_type DROP NOT NULL,
    ALTER COLUMN event_type DROP DEFAULT,
    ALTER COLUMN event_type TYPE VARCHAR(50) USING event_type::text;

ALTER TABLE event_roles
    ALTER COLUMN event_type DROP NOT NULL,
    ALTER COLUMN event_type DROP DEFAULT,
    ALTER COLUMN event_type TYPE VARCHAR(50) USING event_type::text;

ALTER TABLE event_attachments
    ALTER COLUMN event_type DROP NOT NULL,
    ALTER COLUMN event_type DROP DEFAULT,
    ALTER COLUMN event_type TYPE VARCHAR(50) USING event_type::text;

DROP TYPE event_type_enum;
