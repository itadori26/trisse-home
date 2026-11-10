CREATE TABLE appearance_settings (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    theme            TEXT NOT NULL,
    accent_color     TEXT NOT NULL DEFAULT '',
    background_image TEXT NOT NULL DEFAULT '',
    custom_css       TEXT NOT NULL DEFAULT ''
);
