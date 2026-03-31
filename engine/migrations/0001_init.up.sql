CREATE TABLE saga_instance
(
    saga_id            UUID PRIMARY KEY,

    saga_name          TEXT        NOT NULL,
    saga_version       INTEGER     NOT NULL,

    status             TEXT        NOT NULL CHECK (
        status IN (
                   'PENDING',
                   'RUNNING',
                   'COMPLETED',
                   'FAILED',
                   'COMPENSATING',
                   'COMPENSATED',
                   'INCONSISTENT'
            )
        ),

    execution_state    TEXT        NOT NULL CHECK (
        execution_state IN (
                    'RUNNABLE',
                    'WAITING_EVENT'
            )
        ),

    current_step_name  TEXT        NULL,

    idempotency_key    TEXT        NOT NULL,
    correlation_id     TEXT        NULL,

    initial_context    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    runtime_context    JSONB       NOT NULL DEFAULT '{}'::jsonb,

    started_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at        TIMESTAMPTZ NULL,


    last_error_code    TEXT        NULL,
    last_error_message TEXT        NULL,

    context_version    BIGINT      NOT NULL DEFAULT 0,

    next_execution_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_till        TIMESTAMPTZ NULL,
    locked_by          TEXT        NULL
);

CREATE TABLE saga_step
(
    saga_id            UUID        NOT NULL REFERENCES saga_instance (saga_id) ON DELETE CASCADE,
    step_name          TEXT        NOT NULL,

    step_order         INTEGER     NOT NULL,

    status             TEXT        NOT NULL CHECK (
        status IN (
                   'PENDING',
                   'RUNNING',
                   'COMMITTED',
                   'REJECTED',
                   'FAILED',
                   'COMPENSATING',
                   'COMPENSATED',
                   'VERIFYING',
                   'VERIFIED'
            )
        ),

    attempt            INTEGER     NOT NULL DEFAULT 0,

    worker_instance_id TEXT        NULL,

    input_data         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    output_data        JSONB       NOT NULL DEFAULT '{}'::jsonb,
    error_data         JSONB       NULL,

    effect_state       TEXT        NULL CHECK (
        effect_state IN ('NONE', 'UNKNOWN', 'COMMITTED')
        ),

    started_at         TIMESTAMPTZ NULL,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at        TIMESTAMPTZ NULL,

    PRIMARY KEY (saga_id, step_name)
);

CREATE INDEX saga_step_status_idx
    ON saga_step (status);

CREATE INDEX saga_step_saga_order_idx
    ON saga_step (saga_id, step_order);

CREATE UNIQUE INDEX saga_instance_unique_business_key
    ON saga_instance (saga_name, idempotency_key);

CREATE INDEX saga_instance_status_idx
    ON saga_instance (status);

CREATE INDEX saga_instance_started_at_idx
    ON saga_instance (started_at);

CREATE INDEX saga_instance_updated_at_idx
    ON saga_instance (updated_at);