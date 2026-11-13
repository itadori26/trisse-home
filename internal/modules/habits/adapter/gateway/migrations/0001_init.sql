CREATE TABLE habits (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    emoji      TEXT NOT NULL DEFAULT '',
    days       TEXT NOT NULL DEFAULT '', -- weekdays "1,3,5"; empty = every day
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL
);

CREATE TABLE habit_completions (
    habit_id TEXT NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
    day      TEXT NOT NULL, -- YYYY-MM-DD
    PRIMARY KEY (habit_id, day)
);
