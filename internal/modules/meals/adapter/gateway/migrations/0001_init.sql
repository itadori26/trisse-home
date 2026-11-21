-- Meal plan (#682): one row per planned dish. Dishes are personal content
-- (ADR 0018) and are shared with everyone unless their creator narrows it.
CREATE TABLE meals_entries (
    id          TEXT PRIMARY KEY,
    day         TEXT NOT NULL,             -- YYYY-MM-DD
    meal        TEXT NOT NULL DEFAULT '',  -- breakfast, lunch, dinner or ''
    dish        TEXT NOT NULL,
    link        TEXT NOT NULL DEFAULT '',
    ingredients TEXT NOT NULL DEFAULT '',  -- one per line
    owner_id    TEXT NOT NULL,
    visibility  TEXT NOT NULL DEFAULT 'everyone',
    created_at  TEXT NOT NULL
);
CREATE INDEX meals_entries_day ON meals_entries (day);
CREATE INDEX meals_entries_owner ON meals_entries (owner_id);
