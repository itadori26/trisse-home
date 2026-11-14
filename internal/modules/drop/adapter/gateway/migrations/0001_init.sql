-- Drop zone (#101): items expire; file contents are kept apart.
CREATE TABLE drop_items (
    id           TEXT PRIMARY KEY,
    kind         TEXT NOT NULL,
    text         TEXT NOT NULL DEFAULT '',
    file_name    TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL DEFAULT '',
    size         INTEGER NOT NULL,
    created_at   TEXT NOT NULL,
    expires_at   TEXT NOT NULL
);
CREATE INDEX drop_items_expires ON drop_items (expires_at);

CREATE TABLE drop_files (
    item_id TEXT PRIMARY KEY REFERENCES drop_items (id) ON DELETE CASCADE,
    data    BLOB NOT NULL
);
