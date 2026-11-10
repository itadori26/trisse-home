CREATE TABLE search_settings (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    default_provider TEXT NOT NULL
);

CREATE TABLE search_shortcuts (
    keyword      TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    url_template TEXT NOT NULL
);
