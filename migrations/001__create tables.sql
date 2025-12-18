-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE roles
(
    id           UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name         VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL
);

CREATE TABLE users
(
    id                     UUID        DEFAULT uuid_generate_v4() PRIMARY KEY,
    vk_id                  BIGINT UNIQUE,
    first_name             VARCHAR(100) NOT NULL,
    last_name              VARCHAR(100) NOT NULL,
    role_id                UUID,
    photo_url              TEXT,
    geo_available          BOOLEAN     DEFAULT FALSE,
    notification_available BOOLEAN     DEFAULT TRUE,
    created_at             TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT fk_user_role
        FOREIGN KEY (role_id)
            REFERENCES roles (id)
            ON DELETE RESTRICT
);

CREATE TABLE events
(
    id          UUID        DEFAULT uuid_generate_v4() PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    status      VARCHAR(50),
    description TEXT,
    cover       TEXT,
    vk_post_id  BIGINT,
    vk_vote_id  BIGINT,
    lat         NUMERIC(10, 8),
    lon         NUMERIC(11, 8),
    address     TEXT,
    starts_at   TIMESTAMPTZ,
    type        VARCHAR(100),
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE event_participants
(
    id              UUID        DEFAULT uuid_generate_v4() PRIMARY KEY,
    event_id        UUID NOT NULL,
    user_id         UUID NOT NULL,
    is_checked      BOOLEAN     DEFAULT FALSE,
    check_timestamp TIMESTAMPTZ,
    check_lat       NUMERIC(10, 8),
    check_lon       NUMERIC(11, 8),
    check_type      VARCHAR(50),
    check_author    UUID,
    created_at      TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT fk_event_participant_event
        FOREIGN KEY (event_id)
            REFERENCES events (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_event_participant_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE,

    CONSTRAINT uq_event_user UNIQUE (event_id, user_id)
);

CREATE INDEX idx_event_participant_event_id
    ON event_participants (event_id);

CREATE INDEX idx_event_participant_user_id
    ON event_participants (user_id);

CREATE TABLE event_roles
(
    event_id UUID NOT NULL,
    role_id  UUID NOT NULL,

    PRIMARY KEY (event_id, role_id),

    CONSTRAINT fk_event_role_event
        FOREIGN KEY (event_id)
            REFERENCES events (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_event_role_role
        FOREIGN KEY (role_id)
            REFERENCES roles (id)
            ON DELETE CASCADE
);

CREATE TABLE event_orgs
(
    event_id UUID NOT NULL,
    user_id  UUID NOT NULL,

    PRIMARY KEY (event_id, user_id),

    CONSTRAINT fk_event_org_event
        FOREIGN KEY (event_id)
            REFERENCES events (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_event_org_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE TABLE event_attachments
(
    id              UUID        DEFAULT uuid_generate_v4() PRIMARY KEY,
    event_id        UUID NOT NULL,
    attachment_link TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now(),

    CONSTRAINT fk_event_attachment_event
        FOREIGN KEY (event_id)
            REFERENCES events (id)
            ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS event_attachments;
DROP TABLE IF EXISTS event_orgs;
DROP TABLE IF EXISTS event_roles;
DROP TABLE IF EXISTS event_participants;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;