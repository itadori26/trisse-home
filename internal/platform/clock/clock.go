package clock

import "time"

// System is a kernel.Clock backed by the real time.
type System struct{}

// Now returns the current UTC time.
func (System) Now() time.Time { return time.Now().UTC() }
