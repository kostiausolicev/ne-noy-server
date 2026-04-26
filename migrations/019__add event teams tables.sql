-- +goose Up
-- Таблица для хранения команд, участвующих в мероприятиях типа "команды"
CREATE TABLE teams
(
    id         UUID PRIMARY KEY                    NOT NULL DEFAULT uuid_generate_v4(),
    captain_id UUID REFERENCES users (id)          NOT NULL,
    event_id   UUID REFERENCES event_as_teams (id) NOT NULL,
    team_name  VARCHAR(255)                        NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE            NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE            NOT NULL DEFAULT now()
);

-- Таблица для хранения участников команд
CREATE TABLE team_members
(
    id         UUID PRIMARY KEY           NOT NULL DEFAULT uuid_generate_v4(),
    team_id    UUID REFERENCES teams (id) NOT NULL,
    user_id    UUID REFERENCES users (id) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE   NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE   NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE team_members;
DROP TABLE teams;