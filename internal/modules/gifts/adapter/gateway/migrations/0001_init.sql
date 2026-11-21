-- Gift ideas (#684). Personal content (ADR 0018) that is always private:
-- visibility stays 'private' and every query is limited to the owner.
CREATE TABLE gifts_ideas (
    id         TEXT PRIMARY KEY,
    person     TEXT NOT NULL,
    title      TEXT NOT NULL,
    link       TEXT NOT NULL DEFAULT '',
    note       TEXT NOT NULL DEFAULT '',
    bought     INTEGER NOT NULL DEFAULT 0,
    owner_id   TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    created_at TEXT NOT NULL
);
CREATE INDEX gifts_ideas_owner ON gifts_ideas (owner_id);
