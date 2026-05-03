CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Таблица для событий, фиксируемых паттерном transactional outbox
create table if not exists outbox_events
(
    id            uuid primary key,
    aggregatetype text        not null,
    aggregateid   text        not null,
    type          text        not null,
    payload       jsonb not null,
    created_at    timestamptz not null default now()
);

create index if not exists idx_outbox_events_created_at on outbox_events (created_at);

DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_roles
            WHERE rolname = 'debezium'
        ) THEN
            CREATE ROLE debezium WITH LOGIN PASSWORD 'debezium' REPLICATION;
        END IF;
    END
$$;

GRANT CONNECT ON DATABASE warehouse TO debezium;
GRANT USAGE ON SCHEMA public TO debezium;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO debezium;

DROP PUBLICATION IF EXISTS warehouse_outbox_pub;
CREATE PUBLICATION warehouse_outbox_pub FOR TABLE public.outbox_events;