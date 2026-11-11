CREATE TABLE waste_settings (
    id    INTEGER PRIMARY KEY CHECK (id = 1),
    url   TEXT NOT NULL,
    rules TEXT NOT NULL DEFAULT '[]'
);
