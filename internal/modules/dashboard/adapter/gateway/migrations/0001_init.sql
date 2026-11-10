CREATE TABLE dashboard_tabs (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    position INTEGER NOT NULL
);

CREATE TABLE dashboard_widgets (
    id     TEXT PRIMARY KEY,
    tab_id TEXT NOT NULL REFERENCES dashboard_tabs (id) ON DELETE CASCADE,
    kind   TEXT NOT NULL,
    x      INTEGER NOT NULL,
    y      INTEGER NOT NULL,
    w      INTEGER NOT NULL,
    h      INTEGER NOT NULL,
    config TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX dashboard_widgets_tab ON dashboard_widgets (tab_id);
