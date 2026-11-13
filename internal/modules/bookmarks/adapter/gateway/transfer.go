package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets.
var TransferTables = []sqlutil.Table{{Name: "bookmark_folders"}, {Name: "bookmarks"}, {Name: "bookmark_tags"}, {Name: "bookmark_settings"}}
