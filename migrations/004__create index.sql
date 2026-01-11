-- +goose Up

-- Индексы для таблицы users
CREATE INDEX idx_users_role_id ON users(role_id);

-- Составной индекс для аналитических запросов
CREATE INDEX idx_event_participants_event_checked ON event_participants(event_id, is_checked, user_id);
CREATE INDEX idx_event_participants_user_event ON event_participants(user_id, event_id, is_checked);

CREATE INDEX idx_event_orgs_user_id ON event_orgs(user_id);

-- Индексы для таблицы event_attachments
CREATE INDEX idx_event_attachments_event_id ON event_attachments(event_id);
CREATE INDEX idx_event_attachments_created_at ON event_attachments(created_at DESC);

-- Оптимизация для JOIN запросов
CREATE INDEX idx_users_id_with_name ON users(id, vk_id, first_name, last_name, photo_url);

-- +goose Down

DROP INDEX IF EXISTS idx_users_role_id;
DROP INDEX IF EXISTS idx_users_name_search;
DROP INDEX IF EXISTS idx_events_status;
DROP INDEX IF EXISTS idx_events_starts_at;
DROP INDEX IF EXISTS idx_events_status_starts_at;
DROP INDEX IF EXISTS idx_event_participants_event_checked;
DROP INDEX IF EXISTS idx_event_participants_user_event;
DROP INDEX IF EXISTS idx_event_orgs_user_id;
DROP INDEX IF EXISTS idx_event_attachments_event_id;
DROP INDEX IF EXISTS idx_event_attachments_created_at;
DROP INDEX IF EXISTS idx_users_id_with_name;
DROP INDEX IF EXISTS idx_events_future_active;
