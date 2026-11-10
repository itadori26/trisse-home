CREATE TABLE todos (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    due          TEXT,
    position     INTEGER NOT NULL,
    created_at   TEXT NOT NULL,
    completed_at TEXT
);
