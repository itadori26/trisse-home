CREATE TABLE services_history (
    service_id TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    at         TEXT NOT NULL,
    latency_ms REAL NOT NULL,
    online     INTEGER NOT NULL,
    PRIMARY KEY (service_id, at)
);

CREATE INDEX services_history_at ON services_history (at);
