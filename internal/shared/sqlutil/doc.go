// Package sqlutil contains helpers for SQLite gateways: time encoding and
// detection of constraint violations.
//
// Ring: Interface Adapters. Only depends on the standard library so gateways
// stay independent of the concrete SQLite driver.
package sqlutil
