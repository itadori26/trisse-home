CREATE TABLE adguard_settings (
    id                 INTEGER PRIMARY KEY CHECK (id = 1),
    url                TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL
);
