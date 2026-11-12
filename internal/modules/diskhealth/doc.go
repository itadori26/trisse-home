// Package diskhealth is the S.M.A.R.T. disk health module (#59). It reads
// the smartctl JSON written by scripts/smart-export.sh on the host into a
// read-only mounted folder; the container never gets device access.
//
// Ring: module wiring.
package diskhealth
