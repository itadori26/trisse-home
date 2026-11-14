-- #153: the dates come from an ICS URL or an uploaded file. Existing rows
-- keep their URL source.
ALTER TABLE waste_settings ADD COLUMN source TEXT NOT NULL DEFAULT 'url';
ALTER TABLE waste_settings ADD COLUMN file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE waste_settings ADD COLUMN file_uploaded_at TEXT NOT NULL DEFAULT '';
ALTER TABLE waste_settings ADD COLUMN file_events INTEGER NOT NULL DEFAULT 0;
ALTER TABLE waste_settings ADD COLUMN file_from TEXT NOT NULL DEFAULT '';
ALTER TABLE waste_settings ADD COLUMN file_to TEXT NOT NULL DEFAULT '';

-- The uploaded ICS file (text, at most 1 MB); one row. Stored as TEXT so
-- data exports carry it unchanged.
CREATE TABLE waste_calendar_file (
    id      INTEGER PRIMARY KEY CHECK (id = 1),
    content TEXT NOT NULL
);
