package kernel

import "regexp"

// Client-generated IDs (#80): an offline client creates items with its own
// ID so that replaying the same request later is idempotent. They must look
// like the server's IDs: 8–64 letters, digits, '-' or '_'.
var clientID = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

// ValidClientID reports whether id is acceptable as a client-generated ID.
func ValidClientID(id string) bool { return clientID.MatchString(id) }

// ErrInvalidClientID is returned for malformed client-generated IDs.
var ErrInvalidClientID = NewValidationError("id", "The ID must have 8 to 64 letters, digits, '-' or '_'.")
