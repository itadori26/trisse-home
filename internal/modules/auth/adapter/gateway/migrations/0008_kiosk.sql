-- Kiosk displays (#50): a third account role. SQLite cannot change a CHECK
-- constraint, so auth_accounts is rebuilt with the same columns.
CREATE TABLE auth_accounts_v2 (
    id               TEXT PRIMARY KEY,
    role             TEXT NOT NULL CHECK (role IN ('owner', 'guest', 'kiosk')),
    name             TEXT NOT NULL UNIQUE,
    password_hash    TEXT NOT NULL,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    totp_secret_enc  TEXT NOT NULL DEFAULT '',
    totp_pending_enc TEXT NOT NULL DEFAULT '',
    totp_last_step   INTEGER NOT NULL DEFAULT 0,
    disabled         INTEGER NOT NULL DEFAULT 0,
    tabs             TEXT NOT NULL DEFAULT '[]'
);
INSERT INTO auth_accounts_v2 (id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs)
SELECT id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs FROM auth_accounts;
DROP TABLE auth_accounts;
ALTER TABLE auth_accounts_v2 RENAME TO auth_accounts;
CREATE UNIQUE INDEX auth_accounts_one_owner ON auth_accounts (role) WHERE role = 'owner';

-- The secret link (hashed) and the display options of each kiosk account.
CREATE TABLE auth_kiosks (
    account_id     TEXT PRIMARY KEY,
    token_hash     TEXT NOT NULL UNIQUE,
    rotate_seconds INTEGER NOT NULL DEFAULT 0,
    night_from     TEXT NOT NULL DEFAULT '',
    night_to       TEXT NOT NULL DEFAULT '',
    last_used_at   TEXT,
    last_used_ip   TEXT NOT NULL DEFAULT ''
);
