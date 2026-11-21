-- Sample tabs (#670): the example tabs created at setup are marked, so the
-- template choice on first start can replace them while they are untouched.
ALTER TABLE dashboard_tabs ADD COLUMN sample INTEGER NOT NULL DEFAULT 0;
