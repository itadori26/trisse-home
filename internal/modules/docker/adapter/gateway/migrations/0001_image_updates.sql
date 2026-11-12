-- Image update hints (#107).
CREATE TABLE docker_image_updates (
    ref          TEXT PRIMARY KEY,
    local_digest TEXT NOT NULL,
    status       TEXT NOT NULL,
    newer_tag    TEXT NOT NULL DEFAULT '',
    release_url  TEXT NOT NULL DEFAULT '',
    checked_at   TEXT NOT NULL,
    error        TEXT NOT NULL DEFAULT ''
);

CREATE TABLE docker_update_settings (
    id          INTEGER PRIMARY KEY CHECK (id = 1),
    enabled     INTEGER NOT NULL,
    semver_tags INTEGER NOT NULL
);

CREATE TABLE docker_registry_credentials (
    host               TEXT PRIMARY KEY,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL
);
