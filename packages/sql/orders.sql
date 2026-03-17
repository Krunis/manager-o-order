CREATE TABLE IF NOT EXISTS orders(
    idemp_key PRIMARY KEY,
    employee_id TEXT,
    department_id TEXT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    confirmation_employee_id TEXT,

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
    sent_at TIMESTAMPTZ,
)