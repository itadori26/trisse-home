-- Chore plan (#683). Chores are personal content (ADR 0018), shared with
-- everyone unless their creator narrows it; everyone who sees a chore may
-- mark it done, only the creator changes it.
CREATE TABLE chores_items (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    every_days   INTEGER NOT NULL,
    due          TEXT NOT NULL,             -- YYYY-MM-DD
    assignees    TEXT NOT NULL DEFAULT '',  -- one name per line, in turn order
    turn         INTEGER NOT NULL DEFAULT 0,
    last_done_at TEXT NOT NULL DEFAULT '',
    last_done_by TEXT NOT NULL DEFAULT '',
    owner_id     TEXT NOT NULL,
    visibility   TEXT NOT NULL DEFAULT 'everyone',
    created_at   TEXT NOT NULL
);
CREATE INDEX chores_items_owner ON chores_items (owner_id);
