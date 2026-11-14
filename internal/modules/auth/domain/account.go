package domain

import (
	"regexp"
	"slices"
	"time"

	"trisse-home/internal/shared/kernel"
)

// OwnerID identifies the owner account in sessions and events.
const OwnerID = "owner"

// Role decides what an account may do (ADR 0012).
type Role string

// Roles.
const (
	// RoleOwner may do everything; there is exactly one owner.
	RoleOwner Role = "owner"
	// RoleGuest may only read the tabs shared with it.
	RoleGuest Role = "guest"
	// RoleKiosk is a wall display: reads its tabs, signs in with a link (#50).
	RoleKiosk Role = "kiosk"
)

// MaxGuests bounds the number of guest accounts.
const MaxGuests = 20

var guestNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{1,31}$`)

// Account is a user of the application: the owner, or a read-only guest.
type Account struct {
	id           string
	role         Role
	name         string
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
	disabled     bool
	tabs         []string
	totp         TOTP
}

// NewAccount creates the owner account from an already computed hash.
func NewAccount(passwordHash string, now time.Time) (Account, error) {
	if passwordHash == "" {
		return Account{}, kernel.NewValidationError("passwordHash", "must not be empty")
	}
	return Account{id: OwnerID, role: RoleOwner, name: OwnerID, passwordHash: passwordHash, createdAt: now, updatedAt: now}, nil
}

// NewGuest creates a guest account that may read the given tabs.
func NewGuest(id, name, passwordHash string, tabs []string, now time.Time) (Account, error) {
	if !guestNamePattern.MatchString(name) || name == OwnerID {
		return Account{}, kernel.NewValidationError("name", "The name must be 2–32 lowercase letters, digits, dots, dashes or underscores.")
	}
	if id == "" || id == OwnerID || passwordHash == "" {
		return Account{}, kernel.NewValidationError("guest", "id and password are required")
	}
	return Account{id: id, role: RoleGuest, name: name, passwordHash: passwordHash, createdAt: now, updatedAt: now, tabs: normaliseTabs(tabs)}, nil
}

// RestoreAccount rebuilds an account from persisted state.
func RestoreAccount(id string, role Role, name, passwordHash string, createdAt, updatedAt time.Time, disabled bool, tabs []string) Account {
	return Account{id: id, role: role, name: name, passwordHash: passwordHash, createdAt: createdAt, updatedAt: updatedAt, disabled: disabled, tabs: tabs}
}

func normaliseTabs(tabs []string) []string {
	out := []string{}
	for _, t := range tabs {
		if t != "" && !slices.Contains(out, t) {
			out = append(out, t)
		}
	}
	return out
}

// ID returns the account identifier.
func (a Account) ID() string { return a.id }

// Role returns the role.
func (a Account) Role() Role { return a.role }

// Name is the login name of a guest ("owner" for the owner).
func (a Account) Name() string { return a.name }

// IsOwner reports whether this is the owner account.
func (a Account) IsOwner() bool { return a.role == RoleOwner }

// Disabled reports whether a guest may currently not sign in.
func (a Account) Disabled() bool { return a.disabled }

// Tabs are the dashboard tabs a guest may see.
func (a Account) Tabs() []string { return slices.Clone(a.tabs) }

// PasswordHash returns the stored hash.
func (a Account) PasswordHash() string { return a.passwordHash }

// CreatedAt returns when the account was set up.
func (a Account) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns when the account was last changed.
func (a Account) UpdatedAt() time.Time { return a.updatedAt }

// ChangePasswordHash replaces the hash.
func (a *Account) ChangePasswordHash(hash string, now time.Time) error {
	if hash == "" {
		return kernel.NewValidationError("passwordHash", "must not be empty")
	}
	a.passwordHash, a.updatedAt = hash, now
	return nil
}

// SetGuestAccess changes the shared tabs and whether a guest is disabled.
func (a *Account) SetGuestAccess(tabs []string, disabled bool, now time.Time) error {
	if a.IsOwner() {
		return ErrOwnerNotAGuest
	}
	a.tabs, a.disabled, a.updatedAt = normaliseTabs(tabs), disabled, now
	return nil
}

// CanSee reports whether the account may see a dashboard tab.
func (a Account) CanSee(tabID string) bool {
	return a.IsOwner() || slices.Contains(a.tabs, tabID)
}
