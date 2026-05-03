CREATE TABLE IF NOT EXISTS fct_orders
(
    order_id     UUID,
    user_id      UUID,

    status       LowCardinality(String),
    total_amount Decimal(12, 2),
    currency     LowCardinality(String),

    created_at   DateTime64(3, 'UTC'),
    updated_at   DateTime64(3, 'UTC'),

    version      UInt32,

    loaded_at    DateTime64(3, 'UTC') DEFAULT now64(3)
)
    ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(created_at)
ORDER BY (order_id);