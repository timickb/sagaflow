-- таблица, из которой читает CDC (Debezium)
create table outbox_events
(
    id            uuid primary key,
    aggregatetype text        not null,
    aggregateid   text        not null,
    type          text        not null,
    payload       jsonb       not null,
    created_at    timestamptz not null default now(),
    headers       jsonb       null,
    partition_key text        null
);

create index idx_outbox_event_created_at on outbox_events (created_at);