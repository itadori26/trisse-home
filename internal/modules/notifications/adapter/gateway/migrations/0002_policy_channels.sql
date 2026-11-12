-- Severities, quiet hours (#69) and additional channels (#87).
ALTER TABLE notification_settings ADD COLUMN policy TEXT NOT NULL DEFAULT '{}';

CREATE TABLE notification_channels (
    id         TEXT PRIMARY KEY,
    kind       TEXT NOT NULL,
    name       TEXT NOT NULL,
    enabled    INTEGER NOT NULL,
    config     TEXT NOT NULL,          -- JSON without the secret
    secret_enc TEXT NOT NULL DEFAULT ''
);

CREATE TABLE notification_maintenance (
    id     TEXT PRIMARY KEY,
    scope  TEXT NOT NULL,              -- JSON array
    starts TEXT NOT NULL,
    ends   TEXT NOT NULL,
    note   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE notification_pending (
    seq     INTEGER PRIMARY KEY AUTOINCREMENT,
    at      TEXT NOT NULL,
    message TEXT NOT NULL              -- JSON
);
