-- Optimistic concurrency (#80): every change increments the version; a
-- client that edited an older version gets 409 Conflict.
ALTER TABLE notes ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
