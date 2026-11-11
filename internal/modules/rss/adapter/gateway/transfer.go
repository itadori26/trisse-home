package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets. Read state and cached items are not exported; feeds are fetched again.
var TransferTables = []sqlutil.Table{{Name: "rss_feeds", Exclude: []string{"last_fetched_at", "last_error"}}}
