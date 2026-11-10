// Package sqlutil contains helpers for SQLite gateways: time encoding and
// detection of constraint violations.
//
// Ring: Interface Adapters. Only depends on the standard library so gateways
// stay independent of the concrete SQLite driver.
package sqlutil

import (
	"fmt"
	"strings"
	"time"
)

// timeLayout sorts lexically in chronological order, so range queries on
// TEXT columns work as expected.
const timeLayout = "2006-01-02T15:04:05.000000000Z07:00"

// FormatTime encodes t as UTC text for storage.
func FormatTime(t time.Time) string { return t.UTC().Format(timeLayout) }

// ParseTime decodes a value written by FormatTime.
func ParseTime(s string) (time.Time, error) {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stored time %q: %w", s, err)
	}
	return t, nil
}

// FormatNullableTime encodes t, or returns nil for the zero time.
func FormatNullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return FormatTime(t)
}

// ParseNullableTime decodes a nullable time column.
func ParseNullableTime(s *string) (time.Time, error) {
	if s == nil || *s == "" {
		return time.Time{}, nil
	}
	return ParseTime(*s)
}

// IsUniqueViolation reports whether err is a UNIQUE or PRIMARY KEY constraint
// violation. Matching the message keeps gateways free of driver imports.
func IsUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "PRIMARY KEY constraint failed"))
}

// BoolToInt converts a bool to SQLite's integer representation.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
