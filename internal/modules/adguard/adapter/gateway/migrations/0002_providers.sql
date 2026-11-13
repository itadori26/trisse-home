-- Pi-hole as a second DNS filter (#83): one connection per provider.
CREATE TABLE adguard_connections (
    provider           TEXT PRIMARY KEY CHECK (provider IN ('adguard', 'pihole')),
    url                TEXT NOT NULL,
    username           TEXT NOT NULL DEFAULT '',
    password_encrypted TEXT NOT NULL
);

INSERT INTO adguard_connections (provider, url, username, password_encrypted)
SELECT 'adguard', url, username, password_encrypted FROM adguard_settings;

DROP TABLE adguard_settings;
