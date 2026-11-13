-- Read later (#99): saved articles with a sanitised reader copy, and their
-- images so the copy stays readable when the source is gone.
CREATE TABLE readlater_articles (
    id       TEXT PRIMARY KEY,
    url      TEXT NOT NULL,
    title    TEXT NOT NULL,
    site     TEXT NOT NULL DEFAULT '',
    excerpt  TEXT NOT NULL DEFAULT '',
    content  TEXT NOT NULL DEFAULT '',
    words    INTEGER NOT NULL DEFAULT 0,
    status   TEXT NOT NULL DEFAULT 'unread',
    saved_at TEXT NOT NULL,
    bytes    INTEGER NOT NULL DEFAULT 0,
    error    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX readlater_articles_url ON readlater_articles (url);

CREATE TABLE readlater_images (
    article_id   TEXT NOT NULL REFERENCES readlater_articles (id) ON DELETE CASCADE,
    n            INTEGER NOT NULL,
    content_type TEXT NOT NULL,
    data         BLOB NOT NULL,
    PRIMARY KEY (article_id, n)
);
