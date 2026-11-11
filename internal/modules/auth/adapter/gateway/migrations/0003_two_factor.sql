-- Optional second factor: encrypted TOTP secrets, replay protection and
-- hashed one-time recovery codes.
ALTER TABLE auth_account ADD COLUMN totp_secret_enc TEXT NOT NULL DEFAULT '';
ALTER TABLE auth_account ADD COLUMN totp_pending_enc TEXT NOT NULL DEFAULT '';
ALTER TABLE auth_account ADD COLUMN totp_last_step INTEGER NOT NULL DEFAULT 0;

CREATE TABLE auth_recovery_codes (
    code_hash TEXT PRIMARY KEY
);
