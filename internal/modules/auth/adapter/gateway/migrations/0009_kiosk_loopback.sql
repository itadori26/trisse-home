-- The display of an appliance (#140) signs in with a kiosk link that is
-- accepted only from the device itself.
ALTER TABLE auth_kiosks ADD COLUMN loopback_only INTEGER NOT NULL DEFAULT 0;
