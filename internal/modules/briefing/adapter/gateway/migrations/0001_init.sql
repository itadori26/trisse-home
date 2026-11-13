-- Briefing settings (#52, #70): section layout as JSON, the daily push
-- schedule and the weather location. One row.
CREATE TABLE briefing_settings (
    id          INTEGER PRIMARY KEY CHECK (id = 1),
    layout      TEXT NOT NULL DEFAULT '[]',
    push_on     INTEGER NOT NULL DEFAULT 0,
    push_minute INTEGER NOT NULL DEFAULT 420,
    push_days   TEXT NOT NULL DEFAULT '',
    skip_empty  INTEGER NOT NULL DEFAULT 1,
    lat         REAL NOT NULL DEFAULT 0,
    lon         REAL NOT NULL DEFAULT 0,
    last_sent   TEXT
);
