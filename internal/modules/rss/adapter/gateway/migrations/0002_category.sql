-- Feed categories (#97): kept from and written to OPML ('' = none).
ALTER TABLE rss_feeds ADD COLUMN category TEXT NOT NULL DEFAULT '';
