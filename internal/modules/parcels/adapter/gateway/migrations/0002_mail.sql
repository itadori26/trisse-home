-- Parcels from shipping e-mails (#166): the read-only mailbox, where the
-- last run stopped, and numbers the user removed (never added again).
ALTER TABLE parcels ADD COLUMN discovered INTEGER NOT NULL DEFAULT 0;

CREATE TABLE parcels_mail (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    enabled          INTEGER NOT NULL DEFAULT 0,
    host             TEXT NOT NULL DEFAULT '',
    port             INTEGER NOT NULL DEFAULT 993,
    username         TEXT NOT NULL DEFAULT '',
    password_enc     TEXT NOT NULL DEFAULT '',
    folder           TEXT NOT NULL DEFAULT 'INBOX',
    interval_minutes INTEGER NOT NULL DEFAULT 15,
    uid_validity     INTEGER NOT NULL DEFAULT 0,
    last_uid         INTEGER NOT NULL DEFAULT 0,
    last_run         TEXT,
    last_error       TEXT NOT NULL DEFAULT '',
    found            INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE parcels_dismissed (
    tracking_number TEXT PRIMARY KEY,
    dismissed_at    TEXT NOT NULL
);
