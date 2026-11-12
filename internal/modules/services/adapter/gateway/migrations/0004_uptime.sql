-- Uptime statistics and incident log (#81).
ALTER TABLE services_history ADD COLUMN maintenance INTEGER NOT NULL DEFAULT 0;

CREATE TABLE services_daily (
    service_id  TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    day         TEXT NOT NULL,
    checks      INTEGER NOT NULL DEFAULT 0,
    online      INTEGER NOT NULL DEFAULT 0,
    maintenance INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (service_id, day)
);

CREATE TABLE services_incidents (
    id         TEXT PRIMARY KEY,
    service_id TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    started_at TEXT NOT NULL,
    ended_at   TEXT,
    cause      TEXT NOT NULL DEFAULT ''
);
CREATE INDEX services_incidents_service ON services_incidents (service_id, started_at);
