ALTER TABLE dashboard_tabs ADD COLUMN background TEXT NOT NULL DEFAULT '';

CREATE TABLE dashboard_templates (
    id      TEXT PRIMARY KEY,
    name    TEXT NOT NULL,
    widgets TEXT NOT NULL
);
