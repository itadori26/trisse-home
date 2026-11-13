CREATE TABLE immich_settings (
    id                INTEGER PRIMARY KEY CHECK (id = 1),
    url               TEXT NOT NULL,
    api_key_encrypted TEXT NOT NULL
);
