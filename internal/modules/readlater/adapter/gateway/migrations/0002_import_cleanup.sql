-- Read later, part 2 (#99): imported links whose reader copy is fetched in
-- the background, when the status last changed (cleanup of archived
-- articles) and the cleanup setting.
ALTER TABLE readlater_articles ADD COLUMN pending INTEGER NOT NULL DEFAULT 0;
ALTER TABLE readlater_articles ADD COLUMN status_at TEXT;
CREATE INDEX readlater_articles_pending ON readlater_articles (pending) WHERE pending = 1;

CREATE TABLE readlater_settings (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    archive_days INTEGER NOT NULL DEFAULT 0
);
