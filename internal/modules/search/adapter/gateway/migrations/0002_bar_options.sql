-- Options of the search field (#221), formerly options of the search widget.
-- capture_prefixes is JSON ({"Todo":"t","Note":"n","Bookmark":"b"}); empty
-- means the defaults.
ALTER TABLE search_settings ADD COLUMN open_in_new_tab INTEGER NOT NULL DEFAULT 0;
ALTER TABLE search_settings ADD COLUMN capture_prefixes TEXT NOT NULL DEFAULT '';
