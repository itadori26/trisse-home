-- Personal API tokens of the owner; only the SHA-256 hash is stored.
CREATE TABLE auth_api_tokens (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    scopes       TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    expires_at   TEXT NOT NULL DEFAULT '',
    last_used_at TEXT NOT NULL DEFAULT '',
    last_used_ip TEXT NOT NULL DEFAULT ''
);
