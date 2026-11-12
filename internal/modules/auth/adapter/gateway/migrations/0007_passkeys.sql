-- Passkeys (WebAuthn credentials) used as a second factor (ADR 0014).
CREATE TABLE auth_passkeys (
    id            TEXT PRIMARY KEY,
    account_id    TEXT NOT NULL,
    credential_id BLOB NOT NULL UNIQUE,
    public_key    BLOB NOT NULL,
    alg           INTEGER NOT NULL,
    sign_count    INTEGER NOT NULL DEFAULT 0,
    name          TEXT NOT NULL,
    created_at    TEXT NOT NULL,
    last_used_at  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX auth_passkeys_account ON auth_passkeys (account_id);
