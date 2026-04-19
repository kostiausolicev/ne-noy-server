-- +goose Up
CREATE TABLE user_activity_records
(
    id           uuid primary key not null default uuid_generate_v4(),
    user_id      uuid references users (id),
    activity_id  uuid references event_as_activities (id),
    activity     varchar(255), -- бег, велосипед
    starts       timestamp without time zone,
    ends         timestamp without time zone,
    param_values jsonb,

    created_at   timestamp without time zone,
    "hash"       text
);

-- +goose Down
DROP TABLE user_activity_records;