-- +goose Up
-- Таблица для хранения вопросов тестов
CREATE TABLE questions
(
    id         UUID PRIMARY KEY         NOT NULL DEFAULT uuid_generate_v4(),
    text       TEXT                     NOT NULL,
    type       VARCHAR(50)              NOT NULL, -- тип вопроса: single_choice, multiple_choice, open_ended
    event_id   UUID                     NOT NULL REFERENCES event_as_tests (id) ON DELETE CASCADE,
    q_order    INT                      NOT NULL, -- порядок вопроса в тесте

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Таблица для хранения вложений к вопросам (например, изображения, видео, документы)
CREATE TABLE question_attachments
(
    id          UUID PRIMARY KEY         NOT NULL DEFAULT uuid_generate_v4(),
    question_id UUID                     NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    type        VARCHAR(50)              NOT NULL, -- тип вложения: image, video, document
    file_name   VARCHAR(255)             NOT NULL,
    url         TEXT                     NOT NULL,

    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Таблица для хранения вариантов ответов на вопросы типа single_choice и multiple_choice
CREATE TABLE answers
(
    id          UUID PRIMARY KEY         NOT NULL DEFAULT uuid_generate_v4(),
    question_id UUID                     NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    is_correct  BOOLEAN                  NOT NULL, -- для single_choice и multiple_choice
    text        TEXT                     NOT NULL,
    points      INT                      NOT NULL DEFAULT 0, -- количество баллов за правильный ответ

    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Таблица для хранения ответов пользователей на вопросы теста
CREATE TABLE user_answers
(
    id          UUID PRIMARY KEY         NOT NULL DEFAULT uuid_generate_v4(),
    user_id     UUID                     NOT NULL, -- ссылка на пользователя
    question_id UUID                     NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    answer_id   UUID                     NOT NULL REFERENCES answers (id) ON DELETE CASCADE,
    text        TEXT,                              -- для open_ended ответов

    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE user_answers;
DROP TABLE answers;
DROP TABLE question_attachments;
DROP TABLE questions;