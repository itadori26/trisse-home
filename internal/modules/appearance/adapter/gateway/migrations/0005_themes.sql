-- Theme gallery (#111): imported themes and the style (scheme, custom CSS,
-- accent color) that applying a theme replaced, so it can be restored.
CREATE TABLE appearance_themes (
    id   TEXT PRIMARY KEY,
    json TEXT NOT NULL
);

CREATE TABLE appearance_css_backup (
    id  INTEGER PRIMARY KEY CHECK (id = 1),
    theme        TEXT NOT NULL DEFAULT '',
    css          TEXT NOT NULL,
    accent_color TEXT NOT NULL DEFAULT ''
);
