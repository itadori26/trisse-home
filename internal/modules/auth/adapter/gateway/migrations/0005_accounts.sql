-- Several accounts: the owner (migrated from auth_account) and read-only
-- guests (ADR 0012). Sessions and recovery codes belong to an account.
CREATE TABLE auth_accounts (
    id               TEXT PRIMARY KEY,
    role             TEXT NOT NULL CHECK (role IN ('owner', 'guest')),
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

CREATE UNIQUE INDEX auth_accounts_one_owner ON auth_accounts (role) WHERE role = 'owner';

INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step)
SELECT 'owner', 'owner', 'owner', password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step
FROM auth_account;

DROP TABLE auth_account;

ALTER TABLE auth_sessions ADD COLUMN account_id TEXT NOT NULL DEFAULT 'owner';
CREATE INDEX auth_sessions_account ON auth_sessions (account_id);

CREATE TABLE auth_recovery_codes_v2 (
    account_id TEXT NOT NULL,
    code_hash  TEXT NOT NULL,
    PRIMARY KEY (account_id, code_hash)
);
INSERT INTO auth_recovery_codes_v2 (account_id, code_hash) SELECT 'owner', code_hash FROM auth_recovery_codes;
DROP TABLE auth_recovery_codes;
ALTER TABLE auth_recovery_codes_v2 RENAME TO auth_recovery_codes;
