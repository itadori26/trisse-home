-- Holidays (#55): the optional ICS link of the school holidays and the
-- loaded periods, cached per year.
CREATE TABLE holidays_settings (
    id         INTEGER PRIMARY KEY CHECK (id = 1),
    school_url TEXT NOT NULL DEFAULT ''
);

CREATE TABLE holidays_school (
    year       INTEGER NOT NULL,
    name       TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date   TEXT NOT NULL,
    url        TEXT NOT NULL,
    loaded_at  TEXT NOT NULL
);
CREATE INDEX holidays_school_year ON holidays_school (year);
