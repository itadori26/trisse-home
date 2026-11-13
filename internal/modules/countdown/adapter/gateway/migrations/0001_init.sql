CREATE TABLE countdown_items (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    emoji      TEXT NOT NULL DEFAULT '',
    day        TEXT NOT NULL,              -- YYYY-MM-DD
    minute     INTEGER NOT NULL DEFAULT -1, -- minutes after midnight, -1 = whole day
    created_at TEXT NOT NULL
);
