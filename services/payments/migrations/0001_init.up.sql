CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Сущности заказов
CREATE TABLE orders
(
    id           UUID                            NOT NULL
        PRIMARY KEY,
    user_id      UUID                            NOT NULL,
    status       TEXT        DEFAULT 'PENDING'   NOT NULL
        CONSTRAINT orders_status_check
            CHECK (status IN ('PENDING', 'PAID', 'CANCELLED', 'REFUNDED')),
    total_amount NUMERIC(12, 2)                  NOT NULL,
    currency     TEXT        DEFAULT 'USD'::TEXT NOT NULL,
    details      JSONB                           NOT NULL DEFAULT '{}'::JSONB,
    version      INTEGER                         NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ DEFAULT NOW()       NOT NULL,
    updated_at   TIMESTAMPTZ DEFAULT NOW()       NOT NULL
);

-- Платежи по заказам
CREATE TABLE payments
(
    id                  UUID                                NOT NULL
        PRIMARY KEY,
    order_id            UUID                                NOT NULL
        CONSTRAINT payments_orders_fk
            REFERENCES public.orders,
    amount              NUMERIC(12, 2)                      NOT NULL,
    currency            TEXT        DEFAULT 'RUB'::TEXT     NOT NULL,
    payment_method      TEXT                                NOT NULL,
    payment_provider    TEXT                                NOT NULL,
    provider_payment_id TEXT,
    status              TEXT        DEFAULT 'PENDING'::TEXT NOT NULL
        CONSTRAINT payments_status_check
            CHECK (status IN ('PENDING', 'AUTHORIZED', 'CAPTURED', 'FAILED', 'REFUNDED', 'PARTIALLY_REFUNDED')),
    captured_at         TIMESTAMPTZ,
    refunded_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW()           NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT NOW()           NOT NULL
);

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

-- Таблица для событий аналитики
CREATE TABLE IF NOT EXISTS domain_outbox_events
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
CREATE INDEX IF NOT EXISTS idx_domain_outbox_events_created_at
    ON domain_outbox_events (created_at);
CREATE INDEX IF NOT EXISTS idx_payments_order_id
    ON payments (order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status
    ON payments (status);

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_roles
                       WHERE rolname = 'debezium') THEN
            CREATE ROLE debezium WITH LOGIN PASSWORD 'debezium' REPLICATION;
        END IF;
    END
$$;

GRANT CONNECT ON DATABASE payments TO debezium;
GRANT USAGE ON SCHEMA public TO debezium;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO debezium;

DROP PUBLICATION IF EXISTS payments_saga_outbox_pub;
CREATE PUBLICATION payments_saga_outbox_pub FOR TABLE public.saga_outbox_events;

DROP PUBLICATION IF EXISTS payments_domain_outbox_pub;
CREATE PUBLICATION payments_domain_outbox_pub FOR TABLE public.domain_outbox_events;

