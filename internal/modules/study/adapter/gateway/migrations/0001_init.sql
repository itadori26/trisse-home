-- Study progress (#57): modules with credits, status and grade, and the
-- credits of the whole degree.
CREATE TABLE study_modules (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    ects     REAL NOT NULL,
    semester INTEGER NOT NULL DEFAULT 0,
    status   TEXT NOT NULL,
    grade    REAL
);

CREATE TABLE study_settings (
    id         INTEGER PRIMARY KEY CHECK (id = 1),
    total_ects REAL NOT NULL
);
