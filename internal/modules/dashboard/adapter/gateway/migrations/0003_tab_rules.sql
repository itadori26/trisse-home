-- Default-tab rules per tab (#71): JSON list of {days, from, to, networks}.
ALTER TABLE dashboard_tabs ADD COLUMN rules TEXT NOT NULL DEFAULT '[]';
