package domain

import (
	"trisse-home/internal/shared/kernel"
)

// Format identifies trisse-home export files.
const Format = "trisse-home-export"

// Version is the current version of the export format. Files of a newer
// version are rejected; older versions stay importable.
const Version = 1

// MaxImportBytes limits the size of an uploaded export.
const MaxImportBytes = 16 << 20

// CheckHeader validates format and version of an export file.
func CheckHeader(format string, version int) error {
	if format != Format {
		return kernel.NewValidationError("format", "This is not a trisse-home export file.")
	}
	if version < 1 || version > Version {
		return kernel.NewValidationErrorf("version", "Export version {version} is not supported by this version of trisse-home.", "version", version)
	}
	return nil
}
