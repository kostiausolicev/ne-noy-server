-- +goose Up
CREATE TABLE IF NOT EXISTS attachments
(
    id         BIGINT PRIMARY KEY,
    filename   VARCHAR(255) NOT NULL,
    url        TEXT         NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE event_attachments DROP COLUMN attachment_link;
ALTER TABLE event_attachments
    ADD COLUMN attachment_id BIGINT;
ALTER TABLE event_attachments
    ADD CONSTRAINT fk_attachment FOREIGN KEY (attachment_id) REFERENCES attachments (id);

-- +goose Down
DROP TABLE IF EXISTS attachments;
ALTER TABLE event_attachments
    DROP COLUMN attachment_id;