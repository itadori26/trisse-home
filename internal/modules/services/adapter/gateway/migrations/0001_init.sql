CREATE TABLE services (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    check_url       TEXT NOT NULL,
    icon            TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    skip_tls_verify INTEGER NOT NULL DEFAULT 0,
    position        INTEGER NOT NULL
);
