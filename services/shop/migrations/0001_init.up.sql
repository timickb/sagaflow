create table if not exists categories
(
    id    uuid not null primary key,
    code  text not null,
    title text not null
);

create table if not exists items
(
    id                   uuid not null primary key,
    category_id          uuid not null,
    warehouse_product_id uuid not null,
    picture_url          text not null,
    title                text not null,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    constraint items_categories_fk
        foreign key (category_id) references categories(id)
);
