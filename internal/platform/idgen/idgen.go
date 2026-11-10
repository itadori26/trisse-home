// Package idgen provides a kernel.IDGenerator based on crypto/rand.
//
// Ring: Frameworks & Drivers.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

const idBytes = 16

// Random generates 128-bit random identifiers encoded as hex.
type Random struct{}

// NewID returns a new random identifier.
func (Random) NewID() string {
	b := make([]byte, idBytes)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error on supported platforms.
	return hex.EncodeToString(b)
}
