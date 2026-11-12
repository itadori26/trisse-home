package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets. The health check history and the
// result of the last discovery sync are not exported.
var TransferTables = []sqlutil.Table{{Name: "services"}, {Name: "services_discovery", Exclude: []string{
	"last_sync_at", "last_error", "last_created", "last_updated", "last_hidden", "last_shown", "last_skipped",
}}}
