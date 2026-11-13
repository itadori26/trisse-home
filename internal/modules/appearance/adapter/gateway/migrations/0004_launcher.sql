-- #189: the greeting is always part of the header (only the name is kept);
-- the launcher content is a new setting.
ALTER TABLE appearance_settings DROP COLUMN greeting;
ALTER TABLE appearance_settings ADD COLUMN launcher TEXT NOT NULL DEFAULT 'both';
