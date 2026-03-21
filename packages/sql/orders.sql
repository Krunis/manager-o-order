CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS orders(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    employee_id TEXT,
    department_id TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    confirmation_employee_id TEXT,

    idemp_key TEXT UNIQUE,

    confirmation_status TEXT,
    delivert_status TEXT,
    notification_status TEXT,
    storage_status TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbox(
    idemp_key TEXT PRIMARY KEY,

    topic TEXT NOT NULL,
    key TEXT,
    payload JSONB NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS saga_states(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    order_id UUID REFERENCES orders(id),
    status TEXT,
    current_step INT,

    payload JSONB,

    error TEXT,
    
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
)