CREATE TABLE speedtest_settings (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    url             TEXT NOT NULL,
    token_encrypted TEXT NOT NULL,
    min_down_mbps   REAL NOT NULL DEFAULT 0,
    min_up_mbps     REAL NOT NULL DEFAULT 0
);
