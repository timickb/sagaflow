CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Сущности заказов
create table orders
(
    id           uuid                            not null
        primary key,
    user_id      uuid                            not null,
    status       text        default 'PENDING'   not null
        constraint orders_status_check
            check (status IN ('PENDING', 'PAID', 'CANCELLED', 'REFUNDED')),
    total_amount numeric(12, 2)                  not null,
    currency     text        default 'USD'::text not null,
    details      jsonb                           not null default '{}'::jsonb,
    version      integer                         not null default 1,
    created_at   timestamptz default now()       not null,
    updated_at   timestamptz default now()       not null
);

-- Платежи по заказам
create table payments
(
    id                  uuid                                not null
        primary key,
    order_id            uuid                                not null
        constraint payments_orders_fk
            references public.orders,
    amount              numeric(12, 2)                      not null,
    currency            text        default 'RUB'::text     not null,
    payment_method      text                                not null,
    payment_provider    text                                not null,
    provider_payment_id text,
    status              text        default 'PENDING'::text not null
        constraint payments_status_check
            check (status IN ('PENDING', 'AUTHORIZED', 'CAPTURED', 'FAILED', 'REFUNDED', 'PARTIALLY_REFUNDED')),
    captured_at         timestamptz,
    refunded_at         timestamptz,
    created_at          timestamptz default now()           not null,
    updated_at          timestamptz default now()           not null
);


-- Таблица для событий, фиксируемых паттерном transactional outbox
create table if not exists outbox_events
(
    id            uuid primary key,
    aggregatetype text        not null,
    aggregateid   text        not null,
    type          text        not null,
    payload       jsonb       not null,
    created_at    timestamptz not null default now()
);

create index if not exists idx_outbox_events_created_at on outbox_events (created_at);
create index if not exists idx_payments_order_id on payments (order_id);
create index if not exists idx_payments_status on payments (status);

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

DROP PUBLICATION IF EXISTS payments_outbox_pub;
CREATE PUBLICATION payments_outbox_pub FOR TABLE public.outbox_events;

