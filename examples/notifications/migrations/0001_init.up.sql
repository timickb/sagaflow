create table if not exists notifications
(
    id          uuid        not null primary key,
    subject     text        not null,
    body        text        not null,
    mail_from   text        not null,
    mail_to     text        not null,

    attempts    integer     not null default 0,
    locked_by   text,
    locked_till timestamptz,
    send_after  timestamptz,
    sent_at     timestamptz,
    created_at  timestamptz not null default now()
);