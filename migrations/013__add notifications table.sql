-- +goose Up
create table notifications
(
    id        uuid primary key not null default uuid_generate_v4(),
    title     varchar(500)     not null,
    fragment  varchar(255),
    sendAt    timestamp        not null default now(),
    userRole  uuid references roles (id),              -- для какой роли отправляется уведомление
    forAll    bool             not null default false, -- для всех пользователей

    createdAt timestamp        not null default now()
);

create table notification_user
(
    id             uuid primary key not null default uuid_generate_v4(),
    notificationid uuid references notifications (id),
    userid         uuid references users (id),
    read           bool             not null default false
);

-- +goose Down
drop table notification_user;
drop table notifications;
