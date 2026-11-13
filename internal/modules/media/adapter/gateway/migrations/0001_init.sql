CREATE TABLE media_settings (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    provider        TEXT NOT NULL,
    url             TEXT NOT NULL,
    token_encrypted TEXT NOT NULL
);
