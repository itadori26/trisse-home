package domain

import "trisse-home/internal/shared/kernel"

// Migration identifies an applied schema migration.
type Migration struct{ Module, Version string }

// BackupInfo is what a backup file contains, as found by the inspector.
type BackupInfo struct {
	// Integrity is the result of PRAGMA integrity_check ("ok" when sound).
	Integrity  string
	Migrations []Migration
}

// Restore errors.
var (
	ErrNotADatabase = kernel.NewPublicError(kernel.ErrInvalid, "The file is not a trisse-home backup.")
	ErrCorrupt      = kernel.NewPublicError(kernel.ErrInvalid, "The backup is damaged (integrity check failed).")
	ErrTooNew       = kernel.NewPublicError(kernel.ErrInvalid,
		"The backup was made by a newer version of trisse-home. Update first, then restore it.")
	ErrRestoreNotAllowed = kernel.NewPublicError(kernel.ErrForbidden, "Only the owner can restore backups in the browser.")
)

// ValidateRestore checks a backup before it may replace the database: it
// must be sound, come from trisse-home (auth schema present) and must not
// contain migrations unknown to the running version. Older backups are
// fine; the missing migrations run on the next start.
func ValidateRestore(b BackupInfo, current []Migration) error {
	if b.Integrity != "ok" {
		return ErrCorrupt
	}
	known := make(map[Migration]bool, len(current))
	for _, m := range current {
		known[m] = true
	}
	hasAuth := false
	for _, m := range b.Migrations {
		if m.Module == "auth" {
			hasAuth = true
		}
		if !known[m] {
			return ErrTooNew
		}
	}
	if !hasAuth {
		return ErrNotADatabase
	}
	return nil
}
