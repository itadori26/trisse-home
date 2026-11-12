package domain

import (
	"regexp"
	"slices"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Policy limits.
const (
	MinInterval  = time.Hour
	MaxRetention = 365
	namePrefix   = "trisse-home-"
	nameSuffix   = ".db"
	nameLayout   = "20060102-150405"
)

// ErrBackupNotFound is returned for unknown backup names.
var ErrBackupNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The backup does not exist.")

var namePattern = regexp.MustCompile(`^trisse-home-\d{8}-\d{6}\.db$`)

// Policy describes how often backups are taken and how many are kept.
type Policy struct {
	Interval  time.Duration
	Retention int
}

// NewPolicy validates a policy.
func NewPolicy(interval time.Duration, retention int) (Policy, error) {
	if interval < MinInterval {
		return Policy{}, kernel.NewValidationError("interval", "The backup interval must be at least one hour.")
	}
	if retention < 1 || retention > MaxRetention {
		return Policy{}, kernel.NewValidationError("retention", "Between 1 and 365 backups can be kept.")
	}
	return Policy{Interval: interval, Retention: retention}, nil
}

// File is a stored backup.
type File struct {
	Name      string
	Size      int64
	CreatedAt time.Time
}

// FileName returns the name of a backup taken at t. Names sort
// chronologically and contain no path separators.
func FileName(t time.Time) string {
	return namePrefix + t.UTC().Format(nameLayout) + nameSuffix
}

// CreatedFromName returns the time encoded in a backup file name.
func CreatedFromName(name string) (time.Time, bool) {
	if !ValidName(name) {
		return time.Time{}, false
	}
	t, err := time.Parse(nameLayout, name[len(namePrefix):len(name)-len(nameSuffix)])
	return t, err == nil
}

// ValidName reports whether name is a backup file name created by FileName.
// It is used to reject path traversal in download requests.
func ValidName(name string) bool { return namePattern.MatchString(name) }

// ToPrune returns the files exceeding the retention, oldest first.
func (p Policy) ToPrune(files []File) []File {
	sorted := slices.Clone(files)
	slices.SortFunc(sorted, func(a, b File) int { return b.CreatedAt.Compare(a.CreatedAt) })
	if len(sorted) <= p.Retention {
		return nil
	}
	old := sorted[p.Retention:]
	slices.Reverse(old)
	return old
}

// Due reports whether a new backup should be taken at now given the latest one.
func (p Policy) Due(latest time.Time, now time.Time) bool {
	// A small tolerance keeps a job running every Interval from skipping a
	// backup because the previous one finished a few seconds late.
	const tolerance = time.Minute
	return latest.IsZero() || now.Sub(latest) >= p.Interval-tolerance
}
