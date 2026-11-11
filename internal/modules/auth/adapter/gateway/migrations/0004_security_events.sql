-- Security event log (kept for 90 days, see domain.EventRetention).
CREATE TABLE auth_security_events (
    id         TEXT PRIMARY KEY,
    at         TEXT NOT NULL,
    type       TEXT NOT NULL,
    account_id TEXT NOT NULL DEFAULT '',
    ip         TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    detail     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX auth_security_events_at ON auth_security_events (at);
CREATE INDEX auth_security_events_device ON auth_security_events (account_id, type, user_agent);
