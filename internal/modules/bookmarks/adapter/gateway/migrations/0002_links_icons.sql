-- Link checks, favicons and outbound settings (#72).
CREATE TABLE bookmark_settings (
    id          INTEGER PRIMARY KEY CHECK (id = 1),
    enrich      INTEGER NOT NULL DEFAULT 0,
    check_links INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE bookmark_links (
    bookmark_id TEXT PRIMARY KEY REFERENCES bookmarks (id) ON DELETE CASCADE,
    state       TEXT NOT NULL,
    code        INTEGER NOT NULL DEFAULT 0,
    target      TEXT NOT NULL DEFAULT '',
    checked_at  TEXT NOT NULL
);

CREATE TABLE bookmark_icons (
    bookmark_id  TEXT PRIMARY KEY REFERENCES bookmarks (id) ON DELETE CASCADE,
    content_type TEXT NOT NULL,
    data         BLOB NOT NULL
);
