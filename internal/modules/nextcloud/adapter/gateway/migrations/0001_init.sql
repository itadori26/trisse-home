CREATE TABLE nextcloud_settings (
    id                     INTEGER PRIMARY KEY CHECK (id = 1),
    url                    TEXT NOT NULL,
    username               TEXT NOT NULL,
    app_password_encrypted TEXT NOT NULL
);
