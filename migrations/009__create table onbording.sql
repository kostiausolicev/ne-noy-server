-- +goose Up
CREATE TABLE onbordings
(
    id         VARCHAR(50) NOT NULL PRIMARY KEY,
    platform   VARCHAR(50) NOT NULL,
    is_active  BOOLEAN     NOT NULL     DEFAULT TRUE,
    data       JSONB       NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_watches_onbordings
(
    user_id      uuid        NOT NULL,
    onboarding_id VARCHAR(50) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (onboarding_id) REFERENCES onbordings (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, onboarding_id)
);

-- +goose Down
DROP TABLE user_watches_onbordings;
DROP TABLE onbordings;