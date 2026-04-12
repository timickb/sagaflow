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
-- Заказы (внешний контекст от shop/warehouse)
-- order_id используется для связи с reservations в warehouse
create table if not exists orders
(
    id           uuid           not null primary key,
    user_id      uuid           not null,
    status       text           not null check (
        status IN ('PENDING', 'PAID', 'CANCELLED', 'REFUNDED')
        )                                default 'PENDING',

    total_amount numeric(12, 2) not null,
    currency     text           not null default 'USD',

    created_at   timestamptz    not null default now(),
    updated_at   timestamptz    not null default now()
);

-- Платежи (привязка к заказу)
create table if not exists payments
(
    id                  uuid           not null primary key,
    order_id            uuid           not null,

    amount              numeric(12, 2) not null,
    currency            text           not null default 'USD',

    payment_method      text           not null, -- 'card', 'bank_transfer', etc.
    payment_provider    text           not null, -- 'stripe', ' YooKassa', etc.
    provider_payment_id text,                    -- external ID from provider

    status              text           not null check (
        status IN ('PENDING', 'AUTHORIZED', 'CAPTURED', 'FAILED', 'REFUNDED', 'PARTIALLY_REFUNDED')
        )                                       default 'PENDING',

    captured_at         timestamptz,
    refunded_at         timestamptz,

    created_at          timestamptz    not null default now(),
    updated_at          timestamptz    not null default now(),

    constraint payments_orders_fk
        foreign key (order_id) references orders (id)
);

-- Журнал транзакций (авторизация, списание, возврат)
create table if not exists payment_transactions
(
    id                      uuid           not null primary key,
    payment_id              uuid           not null,

    transaction_type        text           not null check (
        transaction_type IN ('AUTHORIZE', 'CAPTURE', 'REFUND', 'VOID')
        ),

    amount                  numeric(12, 2) not null,

    provider_transaction_id text,  -- external ID from payment provider
    provider_response       jsonb, -- raw response from provider

    status                  text           not null check (
        status IN ('PENDING', 'SUCCESS', 'FAILED', 'REJECTED')
        ),
    failure_reason          text,

    scenario_instance_id    uuid,  -- for saga deduplication
    step_id                 text,  -- saga step name

    created_at              timestamptz    not null default now(),

    constraint payment_transactions_payments_fk
        foreign key (payment_id) references payments (id)
);

create index if not exists idx_outbox_events_created_at on outbox_events (created_at);
create index if not exists idx_payments_order_id on payments (order_id);
create index if not exists idx_payments_status on payments (status);
create index if not exists idx_payment_transactions_payment_id on payment_transactions (payment_id);
create index if not exists idx_payment_transactions_scenario on payment_transactions (scenario_instance_id, step_id);

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

GRANT CONNECT ON DATABASE payments TO debezium;
GRANT USAGE ON SCHEMA public TO debezium;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO debezium;

DROP PUBLICATION IF EXISTS payments_outbox_pub;
CREATE PUBLICATION payments_outbox_pub FOR TABLE public.outbox_events;

