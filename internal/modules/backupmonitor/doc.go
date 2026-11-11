// Package backupmonitor watches external backup jobs: every source has a
// secret ping URL that its backup script calls after a run, and is reported as
// overdue when no successful run arrives within its maximum age.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package backupmonitor
