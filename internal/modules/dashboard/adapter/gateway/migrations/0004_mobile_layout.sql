-- Phone layout per tab (#92): JSON list of {id, hidden, rows}; NULL = derived from the desktop grid.
ALTER TABLE dashboard_tabs ADD COLUMN mobile_layout TEXT;
