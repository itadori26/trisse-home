CREATE TABLE system_history (
    at     TEXT PRIMARY KEY,
    cpu    REAL NOT NULL,
    memory REAL NOT NULL,
    disk   REAL NOT NULL
);
