package kernel

import "time"

// Clock provides the current time. Use cases depend on it instead of calling
// time.Now so that time-dependent rules stay deterministic in tests.
type Clock interface {
	Now() time.Time
}

// IDGenerator creates new unique entity identifiers.
type IDGenerator interface {
	NewID() string
}
