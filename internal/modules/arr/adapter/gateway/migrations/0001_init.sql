CREATE TABLE arr_instances (
    id                TEXT PRIMARY KEY,
    app               TEXT NOT NULL CHECK (app IN ('sonarr', 'radarr', 'prowlarr')),
    name              TEXT NOT NULL,
    url               TEXT NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    position          INTEGER NOT NULL
);
