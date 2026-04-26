-- +goose Up
-- Мероприятия как тесты
CREATE TABLE event_as_tests
(
    id          uuid primary key            not null default uuid_generate_v4(),
    name        varchar(255)                not null,
    description text,
    cover       text,
    status      varchar(50)                 not null,
    starts_at   timestamp without time zone not null,
    ends_at     timestamp without time zone,
    ext_link_id text, -- ссылка на решение теста на сторонней платформе
    attempts    int                         not null default 1,
    vk_post_id  bigint,

    created_at  timestamp with time zone    not null default now(),
    updated_at  timestamp with time zone    not null default now()
);

-- Мероприятия как опросы
CREATE TABLE event_as_polls
(
    id          uuid primary key            not null default uuid_generate_v4(),
    name        varchar(255)                not null,
    description text,
    cover       text,
    status      varchar(50)                 not null,
    starts_at   timestamp without time zone not null,
    ends_at     timestamp without time zone,
    ext_link_id text, -- ссылка на решение теста на сторонней платформе
    vk_post_id  bigint,

    created_at  timestamp with time zone    not null default now(),
    updated_at  timestamp with time zone    not null default now()
);

-- Мероприятия как команды
CREATE TABLE event_as_teams
(
    id                 uuid primary key            not null default uuid_generate_v4(),
    name               varchar(255)                not null,
    description        text,
    cover              text,
    status             varchar(50)                 not null,
    starts_at          timestamp without time zone not null,
    ends_at            timestamp without time zone,
    teams_constraint   int                         not null default 1,
    teams_cap_min      int, -- минимальное количество участников в команде
    teams_cap_max      int, -- максимальное количество участников в команде
    lat                numeric(10, 8),
    lon                numeric(11, 8),
    address            text,
    additional_address varchar(255),
    vk_post_id         bigint,

    created_at         timestamp with time zone    not null default now(),
    updated_at         timestamp with time zone    not null default now()
);

-- Мероприятия как тренировка
CREATE TABLE event_as_activities
(
    id                   uuid primary key            not null default uuid_generate_v4(),
    name                 varchar(255)                not null,
    description          text,
    cover                text,
    status               varchar(50)                 not null,
    starts_at            timestamp without time zone not null,
    ends_at              timestamp without time zone,
    train_params         jsonb, -- параметры тренировки: км, длительность
    available_activities jsonb, -- доступные виды тренировок: бег, велосипед

    created_at           timestamp with time zone    not null default now(),
    updated_at           timestamp with time zone    not null default now()
);

ALTER TABLE events
    RENAME TO event_as_events;

-- Вью для записей мероприятий
CREATE VIEW events AS
SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, 'as_event' as type
FROM event_as_events e
UNION
SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, 'activity' as type
FROM event_as_activities e
UNION
SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, 'as_team' as type
FROM event_as_teams e
UNION
SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, 'poll' as type
FROM event_as_polls e
UNION
SELECT e.id, e.name, e.status, e.starts_at, e.ends_at, 'as_test' as type
FROM event_as_tests e;

-- Удаляем внешний ключ и добавляем колонку для типа мероприятия
ALTER TABLE event_attachments
    DROP CONSTRAINT fk_event_attachment_event;
ALTER TABLE event_attachments
    ADD COLUMN event_type VARCHAR(50);

-- Удаляем внешний ключ и добавляем колонку для типа мероприятия
ALTER TABLE event_orgs
    DROP CONSTRAINT fk_event_org_event;
ALTER TABLE event_orgs
    ADD COLUMN event_type VARCHAR(50);

-- Удаляем внешний ключ и добавляем колонку для типа мероприятия
ALTER TABLE event_roles
    DROP CONSTRAINT fk_event_role_event;
ALTER TABLE event_roles
    ADD COLUMN event_type VARCHAR(50);

-- +goose Down
DROP VIEW events;
ALTER TABLE event_as_events
    RENAME TO events;
DROP TABLE event_as_activities;
DROP TABLE event_as_teams;
DROP TABLE event_as_polls;
DROP TABLE event_as_tests;

ALTER TABLE event_attachments
    DROP COLUMN event_type;
ALTER TABLE event_orgs
    DROP COLUMN event_type;
ALTER TABLE event_roles
    DROP COLUMN event_type;

ALTER TABLE event_attachments
    ADD CONSTRAINT fk_event_attachment_event FOREIGN KEY (event_id) REFERENCES events (id);
ALTER TABLE event_orgs
    ADD CONSTRAINT fk_event_org_event FOREIGN KEY (event_id) REFERENCES events (id);
ALTER TABLE event_roles
    ADD CONSTRAINT fk_event_role_event FOREIGN KEY (event_id) REFERENCES events (id);