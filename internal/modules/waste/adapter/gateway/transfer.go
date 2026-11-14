package gateway

import "trisse-home/internal/shared/sqlutil"

// TransferTables are the tables included in data exports, parents before
// children; they contain no secrets (the uploaded ICS file is stored as text).
var TransferTables = []sqlutil.Table{{Name: "waste_settings"}, {Name: "waste_calendar_file"}}
