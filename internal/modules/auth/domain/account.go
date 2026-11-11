package domain

import (
	"time"

	"trisse-home/internal/shared/kernel"
)

// Account is the single user of the application, identified only by its
// password hash.
type Account struct {
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
	totp         TOTP
}

// NewAccount creates an account from an already computed password hash.
func NewAccount(passwordHash string, now time.Time) (Account, error) {
	if passwordHash == "" {
		return Account{}, kernel.NewValidationError("passwordHash", "must not be empty")
	}
	return Account{passwordHash: passwordHash, createdAt: now, updatedAt: now}, nil
}

// RestoreAccount rebuilds an account from persisted state.
func RestoreAccount(passwordHash string, createdAt, updatedAt time.Time) Account {
	return Account{passwordHash: passwordHash, createdAt: createdAt, updatedAt: updatedAt}
}

// PasswordHash returns the stored hash.
func (a Account) PasswordHash() string { return a.passwordHash }

// CreatedAt returns when the account was set up.
func (a Account) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns when the password was last changed.
func (a Account) UpdatedAt() time.Time { return a.updatedAt }

// ChangePasswordHash replaces the hash.
func (a *Account) ChangePasswordHash(hash string, now time.Time) error {
	if hash == "" {
		return kernel.NewValidationError("passwordHash", "must not be empty")
	}
	a.passwordHash, a.updatedAt = hash, now
	return nil
}
