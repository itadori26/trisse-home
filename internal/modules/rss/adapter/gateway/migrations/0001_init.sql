CREATE TABLE rss_feeds (
    id              TEXT PRIMARY KEY,
    url             TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    site_url        TEXT NOT NULL DEFAULT '',
    last_fetched_at TEXT,
    last_error      TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL
);

CREATE TABLE rss_items (
    id           TEXT PRIMARY KEY,
    feed_id      TEXT NOT NULL REFERENCES rss_feeds (id) ON DELETE CASCADE,
    guid         TEXT NOT NULL,
    title        TEXT NOT NULL,
    link         TEXT NOT NULL DEFAULT '',
    published_at TEXT NOT NULL,
    read         INTEGER NOT NULL DEFAULT 0,
    UNIQUE (feed_id, guid)
);

CREATE INDEX rss_items_published ON rss_items (published_at DESC);
