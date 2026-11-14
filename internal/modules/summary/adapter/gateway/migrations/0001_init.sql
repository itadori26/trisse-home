-- AI summaries (#75): the local Ollama and the summaries per content hash.
CREATE TABLE summary_settings (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    url             TEXT NOT NULL,
    model           TEXT NOT NULL,
    max_input_chars INTEGER NOT NULL
);

CREATE TABLE summary_cache (
    key        TEXT PRIMARY KEY,
    text       TEXT NOT NULL,
    model      TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX summary_cache_created ON summary_cache (created_at);
