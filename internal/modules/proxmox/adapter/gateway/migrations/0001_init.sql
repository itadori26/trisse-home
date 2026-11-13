CREATE TABLE proxmox_settings (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    url              TEXT NOT NULL,
    token_id         TEXT NOT NULL,
    secret_encrypted TEXT NOT NULL,
    fingerprint      TEXT NOT NULL DEFAULT ''
);
