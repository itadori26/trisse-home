package domain

import (
	"regexp"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Kiosk displays (#50): a wall tablet shows some tabs without a full
// account. A kiosk is an account of role "kiosk" without a password; it
// signs in with a secret link whose token is stored hashed.

// MaxKiosks bounds the number of kiosk displays.
const MaxKiosks = 10

// kioskPassword is the password hash of a kiosk account: not a valid hash,
// so no password can ever match it.
const kioskPassword = "!kiosk"

// NewKiosk creates the account of a kiosk display.
func NewKiosk(id, name string, tabs []string, now time.Time) (Account, error) {
	if !guestNamePattern.MatchString(name) || name == OwnerID {
		return Account{}, kernel.NewValidationError("name", "The name must be 2–32 lowercase letters, digits, dots, dashes or underscores.")
	}
	if id == "" || id == OwnerID {
		return Account{}, kernel.NewValidationError("kiosk", "id is required")
	}
	tabs = normaliseTabs(tabs)
	if len(tabs) == 0 {
		return Account{}, kernel.NewValidationError("tabs", "Choose at least one tab for the display.")
	}
	return Account{id: id, role: RoleKiosk, name: name, passwordHash: kioskPassword, createdAt: now, updatedAt: now, tabs: tabs}, nil
}

// KioskOptions control how a display shows its tabs.
type KioskOptions struct {
	// RotateSeconds switches to the next tab; 0 = no rotation.
	RotateSeconds int
	// NightFrom and NightTo ("HH:MM", local time of the display) dim the
	// screen; both empty = never.
	NightFrom string
	NightTo   string
}

var clockTime = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// NewKioskOptions validates the options.
func NewKioskOptions(rotate int, nightFrom, nightTo string) (KioskOptions, error) {
	o := KioskOptions{RotateSeconds: rotate, NightFrom: strings.TrimSpace(nightFrom), NightTo: strings.TrimSpace(nightTo)}
	if rotate != 0 && (rotate < 10 || rotate > 3600) {
		return KioskOptions{}, kernel.NewValidationError("rotateSeconds", "Rotate every 10 seconds to 1 hour, or never.")
	}
	if (o.NightFrom == "") != (o.NightTo == "") ||
		(o.NightFrom != "" && (!clockTime.MatchString(o.NightFrom) || !clockTime.MatchString(o.NightTo) || o.NightFrom == o.NightTo)) {
		return KioskOptions{}, kernel.NewValidationError("nightFrom", "Night dimming needs a start and an end such as 22:00 and 06:30.")
	}
	return o, nil
}

// Kiosk is the link and display options of a kiosk account.
type Kiosk struct {
	AccountID  string
	TokenHash  TokenHash
	Options    KioskOptions
	LastUsedAt time.Time
	LastUsedIP string
}

// ErrKioskNotFound is returned for unknown kiosk displays.
var ErrKioskNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The kiosk display does not exist.")

// ErrInvalidKioskLink is returned for an unknown, renewed or disabled link.
var ErrInvalidKioskLink = kernel.NewPublicError(kernel.ErrUnauthorized, "This kiosk link is not valid (any more). Ask for a new one.")

// ErrTooManyKiosks is returned at the limit.
var ErrTooManyKiosks = kernel.NewPublicError(kernel.ErrConflict, "At most 10 kiosk displays are supported.")

// KioskMayAccess decides whether a kiosk display may call method on path:
// what a guest may read, its own display options and logging out, but none
// of the account self-service (password, two-factor, passkeys, sessions).
func KioskMayAccess(method, path string, reach GuestReach) bool {
	switch path {
	case "/api/auth/logout":
		return method == "POST"
	case "/api/auth/kiosk":
		return method == "GET" || method == "HEAD"
	}
	for _, p := range guestSelfService {
		if path == p || strings.HasPrefix(path, p+"/") {
			return false
		}
	}
	return GuestMayAccess(method, path, reach)
}
