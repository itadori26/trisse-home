CREATE TABLE system_settings (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    update_check INTEGER NOT NULL DEFAULT 0
);
