CREATE TABLE bookmark_folders (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    position INTEGER NOT NULL
);

CREATE TABLE bookmarks (
    id          TEXT PRIMARY KEY,
    folder_id   TEXT REFERENCES bookmark_folders (id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    url         TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE TABLE bookmark_tags (
    bookmark_id TEXT NOT NULL REFERENCES bookmarks (id) ON DELETE CASCADE,
    tag         TEXT NOT NULL,
    PRIMARY KEY (bookmark_id, tag)
);

CREATE INDEX bookmark_tags_tag ON bookmark_tags (tag);
