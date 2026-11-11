CREATE TABLE backup_sources (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    max_age_seconds INTEGER NOT NULL,
    token_hash      TEXT NOT NULL UNIQUE,
    created_at      TEXT NOT NULL,
    last_success_at TEXT,
    last_failure_at TEXT,
    last_message    TEXT NOT NULL DEFAULT ''
);
