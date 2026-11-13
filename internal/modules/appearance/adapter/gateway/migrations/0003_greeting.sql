-- Greeting line below the top bar (#167).
ALTER TABLE appearance_settings ADD COLUMN greeting INTEGER NOT NULL DEFAULT 0;
ALTER TABLE appearance_settings ADD COLUMN greeting_name TEXT NOT NULL DEFAULT '';
