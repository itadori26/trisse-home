CREATE TABLE notification_settings (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    enabled         INTEGER NOT NULL,
    server_url      TEXT NOT NULL,
    topic           TEXT NOT NULL,
    token_encrypted TEXT NOT NULL DEFAULT '',
    rules           TEXT NOT NULL
);
