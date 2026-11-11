package domain

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits and thresholds.
const (
	// OnlineWindow: a peer is online when its last handshake is newer.
	OnlineWindow = 3 * time.Minute
	// StaleAfter: the export is considered outdated after this age.
	StaleAfter    = 5 * time.Minute
	maxNameLength = 40
)

var publicKeyPattern = regexp.MustCompile(`^[A-Za-z0-9+/]{42}[AEIMQUYcgkosw048]=$`)

// ErrNoExport is returned when the export file does not exist (yet).
var ErrNoExport = kernel.NewPublicError(kernel.ErrNotFound,
	"No WireGuard export found. Run scripts/wireguard-export.sh on the host (see README).")

// ValidPublicKey reports whether s looks like a WireGuard public key.
func ValidPublicKey(s string) bool { return publicKeyPattern.MatchString(s) }

// NormaliseName validates a peer name; empty removes the name.
func NormaliseName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) > maxNameLength {
		return "", kernel.NewValidationError("name", "The name must be at most 40 characters long.")
	}
	return name, nil
}

// Peer is one WireGuard peer.
type Peer struct {
	Interface       string
	PublicKey       string
	Name            string
	Endpoint        string
	AllowedIPs      string
	LatestHandshake time.Time
	ReceivedBytes   int64
	SentBytes       int64
}

// Online reports whether the peer had a handshake within OnlineWindow.
func (p Peer) Online(now time.Time) bool {
	return !p.LatestHandshake.IsZero() && now.Sub(p.LatestHandshake) < OnlineWindow
}

// DisplayName is the name or a shortened public key.
func (p Peer) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	if len(p.PublicKey) > 8 {
		return p.PublicKey[:8] + "…"
	}
	return p.PublicKey
}

// SortForDisplay puts online peers first, then the most recent handshake.
func SortForDisplay(peers []Peer, now time.Time) {
	slices.SortStableFunc(peers, func(a, b Peer) int {
		if a.Online(now) != b.Online(now) {
			if a.Online(now) {
				return -1
			}
			return 1
		}
		if c := b.LatestHandshake.Compare(a.LatestHandshake); c != 0 {
			return c
		}
		return cmp.Compare(a.DisplayName(), b.DisplayName())
	})
}
