CREATE TABLE focus_sessions (
    id       TEXT PRIMARY KEY,
    minutes  INTEGER NOT NULL,
    ended_at TEXT NOT NULL
);

CREATE INDEX focus_sessions_ended_at ON focus_sessions (ended_at);
