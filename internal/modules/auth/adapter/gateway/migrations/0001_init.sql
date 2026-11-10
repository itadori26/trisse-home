CREATE TABLE auth_account (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

CREATE TABLE auth_sessions (
    token_hash   TEXT PRIMARY KEY,
    csrf_token   TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    expires_at   TEXT NOT NULL,
    refreshed_at TEXT NOT NULL
);

CREATE INDEX auth_sessions_expires_at ON auth_sessions (expires_at);
