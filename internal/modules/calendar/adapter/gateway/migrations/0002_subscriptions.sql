-- ICS subscriptions (#98). The URL may contain a private token and is
-- stored encrypted like a password.
CREATE TABLE calendar_subscriptions (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    url_encrypted    TEXT NOT NULL,
    color            TEXT NOT NULL,
    refresh_minutes  INTEGER NOT NULL,
    position         INTEGER NOT NULL
);
