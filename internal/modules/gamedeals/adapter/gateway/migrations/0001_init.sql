CREATE TABLE gamedeals_settings (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    steam_id     TEXT NOT NULL DEFAULT '',
    country      TEXT NOT NULL,
    min_discount INTEGER NOT NULL,
    epic         INTEGER NOT NULL DEFAULT 0
);
