CREATE TABLE parcels (
    id                 TEXT PRIMARY KEY,
    tracking_number    TEXT NOT NULL,
    carrier            TEXT NOT NULL,
    label              TEXT NOT NULL DEFAULT '',
    status_code        TEXT NOT NULL DEFAULT 'unknown',
    status_description TEXT NOT NULL DEFAULT '',
    status_location    TEXT NOT NULL DEFAULT '',
    status_at          TEXT,
    estimated_delivery TEXT,
    checked_at         TEXT,
    error              TEXT NOT NULL DEFAULT '',
    created_at         TEXT NOT NULL
);

CREATE TABLE parcels_settings (
    id                  INTEGER PRIMARY KEY CHECK (id = 1),
    dhl_api_key_enc     TEXT NOT NULL DEFAULT '',
    hide_delivered_days INTEGER NOT NULL DEFAULT 3
);
