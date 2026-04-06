CREATE TABLE IF NOT EXISTS employees(
    id TEXT PRIMARY KEY,
    department TEXT
);

CREATE TABLE IF NOT EXISTS storage(
    item_id TEXT PRIMARY KEY,
    item_name TEXT,
    count INT
);

CREATE TABLE IF NOT EXISTS reservations(
    id TEXT PRIMARY KEY,
    item_id TEXT REFERENCES storage(item_id),
    count INT
)