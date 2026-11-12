-- Off-site backup target (#46).
CREATE TABLE backup_remote (
    id                 INTEGER PRIMARY KEY CHECK (id = 1),
    url                TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL,
    retention          INTEGER NOT NULL
);
