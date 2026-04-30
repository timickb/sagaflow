erDiagram
    ORDERS ||--o{ PAYMENTS : "has"
    PAYMENTS ||--o{ PAYMENT_TRANSACTIONS : "has"

    ORDERS {
        uuid id PK
        uuid user_id
        text status
        numeric_12_2 total_amount
        text currency
        timestamptz created_at
        timestamptz updated_at
    }

    PAYMENTS {
        uuid id PK
        uuid order_id FK
        numeric_12_2 amount
        text currency
        text payment_method
        text payment_provider
        text provider_payment_id
        text status
        timestamptz captured_at
        timestamptz refunded_at
        timestamptz created_at
        timestamptz updated_at
    }

    PAYMENT_TRANSACTIONS {
        uuid id PK
        uuid payment_id FK
        text transaction_type
        numeric_12_2 amount
        text provider_transaction_id
        jsonb provider_response
        text status
        text failure_reason
        uuid scenario_instance_id
        text step_id
        timestamptz created_at
    }