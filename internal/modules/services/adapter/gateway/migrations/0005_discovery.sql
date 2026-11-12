-- Docker autodiscovery (#78): tiles created from container labels.
ALTER TABLE services ADD COLUMN container TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0;
ALTER TABLE services ADD COLUMN edited INTEGER NOT NULL DEFAULT 0;

CREATE TABLE services_discovery (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    enabled          INTEGER NOT NULL DEFAULT 0,
    interval_minutes INTEGER NOT NULL,
    rule             TEXT NOT NULL,
    last_sync_at     TEXT NOT NULL DEFAULT '',
    last_error       TEXT NOT NULL DEFAULT '',
    last_created     INTEGER NOT NULL DEFAULT 0,
    last_updated     INTEGER NOT NULL DEFAULT 0,
    last_hidden      INTEGER NOT NULL DEFAULT 0,
    last_shown       INTEGER NOT NULL DEFAULT 0,
    last_skipped     INTEGER NOT NULL DEFAULT 0
);
