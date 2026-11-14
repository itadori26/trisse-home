-- Space widget (#64): both sources are off until switched on; the NASA API
-- key is stored encrypted with SECRET_KEY.
CREATE TABLE space_settings (
    id                INTEGER PRIMARY KEY CHECK (id = 1),
    launches          INTEGER NOT NULL DEFAULT 0,
    picture           INTEGER NOT NULL DEFAULT 0,
    nasa_key_encrypted TEXT NOT NULL DEFAULT ''
);
