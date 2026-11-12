-- Heartbeat checks (#100): backup sources become checks of kind "backup";
-- ping URLs (token hashes) stay unchanged.
ALTER TABLE backup_sources ADD COLUMN kind TEXT NOT NULL DEFAULT 'backup';
ALTER TABLE backup_sources ADD COLUMN cron TEXT NOT NULL DEFAULT '';
ALTER TABLE backup_sources ADD COLUMN grace_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE backup_sources ADD COLUMN last_start_at TEXT;
ALTER TABLE backup_sources ADD COLUMN last_duration_ms INTEGER NOT NULL DEFAULT 0;
