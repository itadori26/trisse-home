CREATE TABLE calendar_settings (
    id                 INTEGER PRIMARY KEY CHECK (id = 1),
    calendar_urls      TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL
);
