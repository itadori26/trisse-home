CREATE TABLE timeline_events (
    id       TEXT PRIMARY KEY,
    at       TEXT NOT NULL,
    source   TEXT NOT NULL,
    severity TEXT NOT NULL,
    subject  TEXT NOT NULL DEFAULT '',
    message  TEXT NOT NULL
);
CREATE INDEX timeline_events_at ON timeline_events (at);

CREATE TABLE timeline_settings (
    id             INTEGER PRIMARY KEY CHECK (id = 1),
    retention_days INTEGER NOT NULL
);
