--- физические склады
create table if not exists warehouses
(
    id         uuid        not null primary key,
    code       text        not null unique,
    name       text        not null,
    created_at timestamptz not null default now()
);

-- все товары
create table if not exists products
(
    id         uuid        not null primary key,
    sku        text        not null unique,
    name       text        not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- остатки товаров
create table balances
(
    warehouse_id      uuid        not null,
    product_id        uuid        not null,

    quantity_total    integer     not null,
    quantity_reserved integer     not null default 0,

    version           bigint      not null,
    updated_at        timestamptz not null default now(),

    primary key (warehouse_id, product_id),

    constraint balances_warehouses_fk
        foreign key (warehouse_id) references warehouses (id),
    constraint balances_products_fk
        foreign key (product_id) references products (id)
);

-- журнал перемещений товаров
create table if not exists movements
(
    movement_id          uuid        not null primary key,
    warehouse_id         uuid        not null,
    product_id           uuid        not null,

    movement_type        text        not null check (
        movement_type IN ('INBOUND', 'OUTBOUND', 'RESERVE', 'RELEASE')
        ),

    quantity             integer     not null check (quantity > 0),
    business_ref         text        not null, -- order_id / shipment_id
    scenario_instance_id uuid        not null, -- ключ дедупликации
    step_id              text        not null, -- шаг саги
    created_at           timestamptz not null default now(),

    unique (scenario_instance_id, step_id)
);

-- брони товаров
create table if not exists reservations
(
    id                   uuid        not null primary key,

    order_id             uuid        not null, -- из сервиса payments
    warehouse_id         uuid        not null,
    product_id           uuid        not null,

    quantity             integer     not null check (quantity > 0),
    status               text        not null check (
        status IN ('ACTIVE', 'CONFIRMED', 'CANCELLED')
        ),

    scenario_instance_id uuid        not null,
    created_at           timestamptz not null default now(),

    unique (order_id, product_id)
);

create table if not exists sagaflow_events
(
    id            uuid        not null primary key,
    topic         text        not null,
    key           text        not null, -- saga_id для партиционирования
    payload       jsonb       not null,
    status        text        not null default 'PENDING',
    attempts      integer     not null default 0,
    created_at    timestamptz not null default now(),
    published_at  timestamptz
);

create index if not exists idx_outbox_events_status on sagaflow_events(status);
create index if not exists idx_outbox_events_topic on sagaflow_events(topic);