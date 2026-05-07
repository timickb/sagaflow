CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Таблица для событий оркестратора, фиксируемых паттерном transactional outbox
CREATE TABLE IF NOT EXISTS saga_outbox_events
(
    id             UUID PRIMARY KEY,
    aggregate_type TEXT        NOT NULL,
    aggregate_id   TEXT        NOT NULL,
    event_type     TEXT        NOT NULL,
    payload        JSONB       NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_saga_outbox_events_created_at
    ON saga_outbox_events (created_at);

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
CREATE PUBLICATION warehouse_outbox_pub FOR TABLE public.saga_outbox_events;