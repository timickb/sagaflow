create table if not exists users
(
    id         uuid        not null primary key,
    username   text        not null,
    email      text        not null,
    first_name text,
    last_name  text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists credentials
(
    id         uuid        not null primary key,
    user_id    uuid        not null,
    type       text        not null,
    hash       text        not null,
    salt       text        not null,
    created_at timestamptz not null default now(),

    constraint credentials_users_fk foreign key (user_id) references users (id)
);

create table if not exists user_sessions
(
    id         uuid        not null primary key,
    user_id    uuid        not null,
    token      text        not null,
    rotations  integer     not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    constraint user_sessions_users_fk foreign key (user_id) references users (id)
);