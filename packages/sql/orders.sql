CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS employees(
    id TEXT PRIMARY KEY,
    department TEXT
);

CREATE TABLE IF NOT EXISTS orders(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    employee_id TEXT,
    department_id TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    confirmation_employee_id TEXT REFERENCES employees(id),

    idemp_key TEXT UNIQUE,

    confirmation_status TEXT,
    storage_status TEXT,
    delivery_status TEXT,
    notification_status TEXT,

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
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS storage(
    item_id TEXT PRIMARY KEY,
    item_name TEXT,
    count INT,
    confirmation_type TEXT
);

CREATE TABLE IF NOT EXISTS reservations(
    id TEXT PRIMARY KEY,
    item_id TEXT REFERENCES storage(item_id),
    count INT
);

CREATE TABLE IF NOT EXISTS confirmations(
    id TEXT PRIMARY KEY,
    employee_id TEXT REFERENCES employees(id)
);

CREATE TABLE IF NOT EXISTS delivery(
    order_id UUID PRIMARY KEY REFERENCES orders(id),
    table_number INTEGER
);

