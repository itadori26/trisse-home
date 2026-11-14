-- Endpoints of the generic JSON widget (#48). The header value is a secret
-- (e.g. an API key) and stored encrypted with SECRET_KEY.
CREATE TABLE apiwidget_endpoints (
    id                     TEXT PRIMARY KEY,
    name                   TEXT NOT NULL,
    url                    TEXT NOT NULL,
    header_name            TEXT NOT NULL DEFAULT '',
    header_value_encrypted TEXT NOT NULL DEFAULT '',
    interval_seconds       INTEGER NOT NULL,
    allow_public           INTEGER NOT NULL DEFAULT 0,
    extractions            TEXT NOT NULL DEFAULT '[]',
    position               INTEGER NOT NULL DEFAULT 0
);
