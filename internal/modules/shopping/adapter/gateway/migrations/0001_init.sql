CREATE TABLE shopping_settings (
    id                 INTEGER PRIMARY KEY CHECK (id = 1),
    list_url           TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL
);
