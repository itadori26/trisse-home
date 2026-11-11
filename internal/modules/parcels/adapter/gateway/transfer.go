package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets. The DHL API key (parcels_settings) is a secret and not exported.
var TransferTables = []sqlutil.Table{{Name: "parcels", Exclude: []string{"error", "checked_at"}}}
