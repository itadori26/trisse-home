-- Full-text search over notes (#73): an external-content FTS5 index kept
-- in sync by triggers; diacritics are folded so "buro" finds "Büro".
CREATE VIRTUAL TABLE notes_fts USING fts5(
    title, body,
    content = 'notes', content_rowid = 'rowid',
    tokenize = 'unicode61 remove_diacritics 2'
);

CREATE TRIGGER notes_fts_insert AFTER INSERT ON notes BEGIN
    INSERT INTO notes_fts (rowid, title, body) VALUES (new.rowid, new.title, new.body);
END;

CREATE TRIGGER notes_fts_delete AFTER DELETE ON notes BEGIN
    INSERT INTO notes_fts (notes_fts, rowid, title, body) VALUES ('delete', old.rowid, old.title, old.body);
END;

CREATE TRIGGER notes_fts_update AFTER UPDATE OF title, body ON notes BEGIN
    INSERT INTO notes_fts (notes_fts, rowid, title, body) VALUES ('delete', old.rowid, old.title, old.body);
    INSERT INTO notes_fts (rowid, title, body) VALUES (new.rowid, new.title, new.body);
END;

INSERT INTO notes_fts (notes_fts) VALUES ('rebuild');

-- Earlier versions of notes (#73).
CREATE TABLE note_revisions (
    id         TEXT PRIMARY KEY,
    note_id    TEXT NOT NULL REFERENCES notes (id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX note_revisions_note ON note_revisions (note_id, created_at);
