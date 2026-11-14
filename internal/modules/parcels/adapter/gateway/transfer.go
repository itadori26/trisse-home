package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets. The DHL API key (parcels_settings) and
// the mailbox with its password (parcels_mail) are not exported; removed
// numbers are, so an import does not bring them back (#166).
var TransferTables = []sqlutil.Table{{Name: "parcels", Exclude: []string{"error", "checked_at"}}, {Name: "parcels_dismissed"}}
