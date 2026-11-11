package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets. The health check history is not exported.
var TransferTables = []sqlutil.Table{{Name: "services"}}
