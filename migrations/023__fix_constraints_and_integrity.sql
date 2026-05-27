-- +goose Up

-- =============================================================================
-- Risk 1 (HIGH): Триггерная проверка ссылочной целостности для полиморфных FK
-- Таблицы event_orgs, event_roles, event_attachments потеряли FK на мероприятия
-- в миграции 016 при переходе на мультитипную архитектуру. Декларативный FK
-- невозможен для полиморфных ссылок, поэтому целостность обеспечивается триггером.
-- =============================================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION validate_event_reference()
    RETURNS TRIGGER AS $$
BEGIN
    CASE NEW.event_type::text
        WHEN 'event' THEN
            IF NOT EXISTS (SELECT 1 FROM event_as_events WHERE id = NEW.event_id) THEN
                RAISE EXCEPTION 'event_id % не найден в event_as_events', NEW.event_id;
            END IF;
        WHEN 'test' THEN
            IF NOT EXISTS (SELECT 1 FROM event_as_tests WHERE id = NEW.event_id) THEN
                RAISE EXCEPTION 'event_id % не найден в event_as_tests', NEW.event_id;
            END IF;
        WHEN 'team' THEN
            IF NOT EXISTS (SELECT 1 FROM event_as_teams WHERE id = NEW.event_id) THEN
                RAISE EXCEPTION 'event_id % не найден в event_as_teams', NEW.event_id;
            END IF;
        WHEN 'poll' THEN
            IF NOT EXISTS (SELECT 1 FROM event_as_polls WHERE id = NEW.event_id) THEN
                RAISE EXCEPTION 'event_id % не найден в event_as_polls', NEW.event_id;
            END IF;
        WHEN 'activity' THEN
            IF NOT EXISTS (SELECT 1 FROM event_as_activities WHERE id = NEW.event_id) THEN
                RAISE EXCEPTION 'event_id % не найден в event_as_activities', NEW.event_id;
            END IF;
        ELSE
            RAISE EXCEPTION 'Неизвестный тип мероприятия: %', NEW.event_type;
    END CASE;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_validate_event_ref_orgs
    BEFORE INSERT OR UPDATE OF event_id, event_type
    ON event_orgs
    FOR EACH ROW EXECUTE FUNCTION validate_event_reference();

CREATE TRIGGER trg_validate_event_ref_roles
    BEFORE INSERT OR UPDATE OF event_id, event_type
    ON event_roles
    FOR EACH ROW EXECUTE FUNCTION validate_event_reference();

CREATE TRIGGER trg_validate_event_ref_attachments
    BEFORE INSERT OR UPDATE OF event_id, event_type
    ON event_attachments
    FOR EACH ROW EXECUTE FUNCTION validate_event_reference();

-- =============================================================================
-- Risk 2 (MEDIUM): Явные ON DELETE для teams, team_members, notification_user
-- Без явного ON DELETE попытка удалить пользователя или мероприятие
-- завершается ошибкой (NO ACTION по умолчанию).
-- Имена ограничений — автогенерированные PostgreSQL при создании таблиц.
-- =============================================================================

-- teams.captain_id: CASCADE — удаление пользователя удаляет его команды
-- (captain_id NOT NULL, SET NULL невозможен без снятия ограничения)
ALTER TABLE teams
    DROP CONSTRAINT teams_captain_id_fkey,
    ADD CONSTRAINT fk_teams_captain
        FOREIGN KEY (captain_id) REFERENCES users (id) ON DELETE CASCADE;

-- teams.event_id: CASCADE — удаление мероприятия удаляет привязанные команды
ALTER TABLE teams
    DROP CONSTRAINT teams_event_id_fkey,
    ADD CONSTRAINT fk_teams_event
        FOREIGN KEY (event_id) REFERENCES event_as_teams (id) ON DELETE CASCADE;

-- team_members.team_id: CASCADE — удаление команды удаляет её участников
ALTER TABLE team_members
    DROP CONSTRAINT team_members_team_id_fkey,
    ADD CONSTRAINT fk_team_members_team
        FOREIGN KEY (team_id) REFERENCES teams (id) ON DELETE CASCADE;

-- team_members.user_id: CASCADE — удаление пользователя удаляет его членства в командах
ALTER TABLE team_members
    DROP CONSTRAINT team_members_user_id_fkey,
    ADD CONSTRAINT fk_team_members_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- notification_user.notificationid: CASCADE — удаление уведомления удаляет ссылки на него
ALTER TABLE notification_user
    DROP CONSTRAINT notification_user_notificationid_fkey,
    ADD CONSTRAINT fk_notification_user_notification
        FOREIGN KEY (notificationid) REFERENCES notifications (id) ON DELETE CASCADE;

-- notification_user.userid: CASCADE — удаление пользователя удаляет его уведомления
ALTER TABLE notification_user
    DROP CONSTRAINT notification_user_userid_fkey,
    ADD CONSTRAINT fk_notification_user_user
        FOREIGN KEY (userid) REFERENCES users (id) ON DELETE CASCADE;

-- =============================================================================
-- Risk 3 (MEDIUM): RAISE NOTICE → RAISE EXCEPTION в триггере проверки отметки
-- Прежняя версия выводила предупреждение, но не блокировала некорректное обновление.
-- =============================================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_check_date_trigger() RETURNS TRIGGER AS
$$
DECLARE
    eventEndsAt TIMESTAMP;
BEGIN
    SELECT e.ends_at
    INTO eventEndsAt
    FROM events e
    WHERE e.id = NEW.event_id;

    IF NEW.check_timestamp IS NOT NULL AND NEW.check_timestamp > eventEndsAt THEN
        RAISE EXCEPTION
            'Время отметки (%) не может быть позже даты завершения мероприятия (%)',
            NEW.check_timestamp, eventEndsAt;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- =============================================================================
-- Risk 4 (LOW): UUID surrogate PK для onboardings
-- Natural key VARCHAR(50) делает переименование экранов разрушительным.
-- Добавляем суррогатный UUID-ключ; старый id сохраняется как UNIQUE natural key.
-- Примечание: требует обновления application-кода — запросы по onboarding_id
-- в user_watches_onboardings теперь используют UUID (internal_onboarding_id).
--
-- Имена ограничений сохраняют оригинальные значения с опечаткой "onbordings",
-- так как PostgreSQL не переименовывает constraint при ALTER TABLE ... RENAME TO.
-- =============================================================================

ALTER TABLE onboardings
    ADD COLUMN internal_id UUID NOT NULL DEFAULT uuid_generate_v4();

ALTER TABLE user_watches_onboardings
    ADD COLUMN internal_onboarding_id UUID;

UPDATE user_watches_onboardings uwo
SET internal_onboarding_id = o.internal_id
FROM onboardings o
WHERE uwo.onboarding_id = o.id;

ALTER TABLE user_watches_onboardings
    ALTER COLUMN internal_onboarding_id SET NOT NULL;

-- Снимаем старые ключи (имена с оригинальной опечаткой "onbordings")
ALTER TABLE user_watches_onboardings
    DROP CONSTRAINT user_watches_onbordings_onboarding_id_fkey,
    DROP CONSTRAINT user_watches_onbordings_pkey;

ALTER TABLE onboardings
    DROP CONSTRAINT onbordings_pkey;

-- Устанавливаем новый PK на UUID и UNIQUE на natural key
ALTER TABLE onboardings
    ADD CONSTRAINT pk_onboardings PRIMARY KEY (internal_id),
    ADD CONSTRAINT uq_onboardings_natural_id UNIQUE (id);

-- Восстанавливаем PK и FK в user_watches_onboardings через новый UUID
ALTER TABLE user_watches_onboardings
    ADD CONSTRAINT pk_user_watches_onboardings
        PRIMARY KEY (user_id, internal_onboarding_id),
    ADD CONSTRAINT fk_uwo_onboarding
        FOREIGN KEY (internal_onboarding_id) REFERENCES onboardings (internal_id) ON DELETE CASCADE;

ALTER TABLE user_watches_onboardings
    DROP COLUMN onboarding_id;

-- =============================================================================
-- Risk 5 (LOW): GENERATED BY DEFAULT AS IDENTITY для attachments.id
-- id без DEFAULT назначается приложением — риск коллизий при репликации.
-- BY DEFAULT (не ALWAYS) сохраняет возможность вставки явных значений.
-- =============================================================================

ALTER TABLE attachments
    ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY;

-- Синхронизируем последовательность с максимальным существующим id
SELECT setval(
    pg_get_serial_sequence('attachments', 'id'),
    COALESCE((SELECT MAX(id) FROM attachments), 0) + 1,
    false
);

-- +goose Down

-- Risk 5
ALTER TABLE attachments
    ALTER COLUMN id DROP IDENTITY IF EXISTS;

-- Risk 4
ALTER TABLE user_watches_onboardings
    ADD COLUMN onboarding_id VARCHAR(50);

UPDATE user_watches_onboardings uwo
SET onboarding_id = o.id
FROM onboardings o
WHERE uwo.internal_onboarding_id = o.internal_id;

ALTER TABLE user_watches_onboardings
    ALTER COLUMN onboarding_id SET NOT NULL;

ALTER TABLE user_watches_onboardings
    DROP CONSTRAINT fk_uwo_onboarding,
    DROP CONSTRAINT pk_user_watches_onboardings;

ALTER TABLE onboardings
    DROP CONSTRAINT pk_onboardings,
    DROP CONSTRAINT uq_onboardings_natural_id;

ALTER TABLE onboardings
    ADD CONSTRAINT onbordings_pkey PRIMARY KEY (id);

ALTER TABLE user_watches_onboardings
    ADD CONSTRAINT user_watches_onbordings_pkey
        PRIMARY KEY (user_id, onboarding_id),
    ADD CONSTRAINT user_watches_onbordings_onboarding_id_fkey
        FOREIGN KEY (onboarding_id) REFERENCES onboardings (id) ON DELETE CASCADE;

ALTER TABLE user_watches_onboardings
    DROP COLUMN internal_onboarding_id;

ALTER TABLE onboardings
    DROP COLUMN internal_id;

-- Risk 3
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION check_check_date_trigger() RETURNS TRIGGER AS
$$
DECLARE
    eventEndsAt TIMESTAMP;
BEGIN
    SELECT e.ends_at
    INTO eventEndsAt
    FROM events e
    WHERE e.id = NEW.event_id;

    IF NEW.check_timestamp IS NOT NULL AND NEW.check_timestamp > eventEndsAt THEN
        RAISE NOTICE 'Время отметки на мероприятии не может быть позже даты завершения мероприятия';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Risk 2
ALTER TABLE notification_user
    DROP CONSTRAINT fk_notification_user_user,
    ADD CONSTRAINT notification_user_userid_fkey
        FOREIGN KEY (userid) REFERENCES users (id);

ALTER TABLE notification_user
    DROP CONSTRAINT fk_notification_user_notification,
    ADD CONSTRAINT notification_user_notificationid_fkey
        FOREIGN KEY (notificationid) REFERENCES notifications (id);

ALTER TABLE team_members
    DROP CONSTRAINT fk_team_members_user,
    ADD CONSTRAINT team_members_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id);

ALTER TABLE team_members
    DROP CONSTRAINT fk_team_members_team,
    ADD CONSTRAINT team_members_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES teams (id);

ALTER TABLE teams
    DROP CONSTRAINT fk_teams_event,
    ADD CONSTRAINT teams_event_id_fkey
        FOREIGN KEY (event_id) REFERENCES event_as_teams (id);

ALTER TABLE teams
    DROP CONSTRAINT fk_teams_captain,
    ADD CONSTRAINT teams_captain_id_fkey
        FOREIGN KEY (captain_id) REFERENCES users (id);

-- Risk 1
DROP TRIGGER IF EXISTS trg_validate_event_ref_attachments ON event_attachments;
DROP TRIGGER IF EXISTS trg_validate_event_ref_roles ON event_roles;
DROP TRIGGER IF EXISTS trg_validate_event_ref_orgs ON event_orgs;
DROP FUNCTION IF EXISTS validate_event_reference();
