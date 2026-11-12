package domain

import (
	"slices"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// RecoveryCodeCount is the number of recovery codes issued when 2FA is enabled.
const RecoveryCodeCount = 10

// Two-factor errors.
var (
	ErrTOTPAlreadyEnabled = kernel.NewPublicError(kernel.ErrConflict, "Two-factor authentication is already enabled.")
	ErrTOTPNotEnabled     = kernel.NewPublicError(kernel.ErrConflict, "Two-factor authentication is not enabled.")
	ErrTOTPNoEnrollment   = kernel.NewPublicError(kernel.ErrConflict, "Start the two-factor setup first.")
	ErrInvalidTOTPCode    = kernel.NewPublicError(kernel.ErrUnauthorized, "The authentication code is incorrect.")
	// ErrTOTPCodeRequired asks the client to send the second factor.
	ErrTOTPCodeRequired = kernel.NewValidationError("totpCode", "Please enter the code from your authenticator app or a recovery code.")
)

// TOTP is the two-factor state of the account. The secret is plain text in
// the domain; gateways encrypt it at rest.
type TOTP struct {
	Secret        string
	PendingSecret string
	// LastStep is the time step of the last accepted code; codes of the same
	// or an earlier step are rejected to prevent replay.
	LastStep      int64
	RecoveryCodes []TokenHash
}

// TwoFactor returns the two-factor state.
func (a Account) TwoFactor() TOTP { return a.totp }

// RestoreTwoFactor sets the persisted two-factor state.
func (a *Account) RestoreTwoFactor(t TOTP) { a.totp = t }

// TOTPEnabled reports whether a second factor is required at login.
func (a Account) TOTPEnabled() bool { return a.totp.Secret != "" }

// StartTOTPEnrollment stores a new, not yet confirmed secret.
func (a *Account) StartTOTPEnrollment(secret string) error {
	if a.TOTPEnabled() {
		return ErrTOTPAlreadyEnabled
	}
	a.totp.PendingSecret = secret
	return nil
}

// ConfirmTOTP activates the pending secret after the user proved possession
// with a code of time step step, and stores the hashes of recovery codes.
func (a *Account) ConfirmTOTP(step int64, recoveryCodes []string) error {
	if a.totp.PendingSecret == "" {
		return ErrTOTPNoEnrollment
	}
	a.totp = TOTP{Secret: a.totp.PendingSecret, LastStep: step}
	for _, c := range recoveryCodes {
		a.totp.RecoveryCodes = append(a.totp.RecoveryCodes, hashRecoveryCode(c))
	}
	return nil
}

// AcceptStep records a successful TOTP code; reused codes are rejected.
func (a *Account) AcceptStep(step int64) error {
	if step <= a.totp.LastStep {
		return ErrInvalidTOTPCode
	}
	a.totp.LastStep = step
	return nil
}

// UseRecoveryCode consumes a recovery code; each code works once.
func (a *Account) UseRecoveryCode(code string) error {
	h := hashRecoveryCode(code)
	i := slices.Index(a.totp.RecoveryCodes, h)
	if i < 0 {
		return ErrInvalidTOTPCode
	}
	a.totp.RecoveryCodes = slices.Delete(a.totp.RecoveryCodes, i, i+1)
	return nil
}

// DisableTOTP removes the second factor.
func (a *Account) DisableTOTP() error {
	if !a.TOTPEnabled() {
		return ErrTOTPNotEnabled
	}
	a.totp = TOTP{}
	return nil
}

// IsRecoveryCodeFormat reports whether input looks like a recovery code
// rather than a 6-digit TOTP code.
func IsRecoveryCodeFormat(input string) bool {
	return len(normaliseRecoveryCode(input)) == 8
}

func normaliseRecoveryCode(code string) string {
	return strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(code)))
}

func hashRecoveryCode(code string) TokenHash { return HashToken(normaliseRecoveryCode(code)) }

// IssueRecoveryCodes replaces the recovery codes, e.g. when the first
// passkey is registered on an account without TOTP.
func (a *Account) IssueRecoveryCodes(codes []string) {
	a.totp.RecoveryCodes = nil
	for _, c := range codes {
		a.totp.RecoveryCodes = append(a.totp.RecoveryCodes, hashRecoveryCode(c))
	}
}

// HasRecoveryCodes reports whether unused recovery codes exist.
func (a Account) HasRecoveryCodes() bool { return len(a.totp.RecoveryCodes) > 0 }
