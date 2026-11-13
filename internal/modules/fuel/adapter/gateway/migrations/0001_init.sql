CREATE TABLE fuel_settings (
    id                INTEGER PRIMARY KEY CHECK (id = 1),
    api_key_encrypted TEXT NOT NULL,
    lat               REAL NOT NULL,
    lon               REAL NOT NULL,
    place             TEXT NOT NULL DEFAULT '',
    radius_km         REAL NOT NULL,
    fuel              TEXT NOT NULL,
    favourites        TEXT NOT NULL DEFAULT '[]'
);
