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
    created_at           timestamptz not null default now()
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

    scenario_instance_id uuid,
    created_at           timestamptz not null default now(),

    unique (order_id, product_id)
);

-- Тестовые данные
-- warehouses
insert into warehouses (id, code, name, created_at)
values
    ('11111111-1111-1111-1111-111111111111', 'MSK-01', 'Основной склад Москва', now()),
    ('22222222-2222-2222-2222-222222222222', 'SPB-01', 'Склад Санкт-Петербург', now()),
    ('33333333-3333-3333-3333-333333333333', 'KZN-01', 'Склад Казань', now());

-- products
insert into products (id, sku, name, created_at, updated_at)
values
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'SKU-LAPTOP-001', 'Ноутбук 14"', now(), now()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'SKU-MOUSE-001', 'Беспроводная мышь', now(), now()),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'SKU-KEYBOARD-001', 'Механическая клавиатура', now(), now()),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'SKU-MONITOR-001', 'Монитор 27"', now(), now());

-- balances
insert into balances (
    warehouse_id,
    product_id,
    quantity_total,
    quantity_reserved,
    version,
    updated_at
)
values
    (
        '11111111-1111-1111-1111-111111111111',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        20,
        2,
        1,
        now()
    ),
    (
        '11111111-1111-1111-1111-111111111111',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        100,
        5,
        1,
        now()
    ),
    (
        '11111111-1111-1111-1111-111111111111',
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        40,
        0,
        1,
        now()
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        10,
        1,
        1,
        now()
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'dddddddd-dddd-dddd-dddd-dddddddddddd',
        15,
        0,
        1,
        now()
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        50,
        0,
        1,
        now()
    );
