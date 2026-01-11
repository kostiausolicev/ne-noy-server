-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE queue_events
(
    id          UUID      DEFAULT uuid_generate_v4() PRIMARY KEY,

    post_id     BIGINT NOT NULL UNIQUE,
    text        TEXT,
    lat         DECIMAL(10, 8),
    lon         DECIMAL(11, 8),
    address     TEXT,
    attachments JSONB,
    poll        JSONB,
    photos      JSONB,

    created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE events
    ADD COLUMN vk_poll_answer_id BIGINT UNIQUE;

-- +goose Down

DROP TABLE queue_events;
ALTER TABLE events
    DROP COLUMN vk_poll_answer_id;
