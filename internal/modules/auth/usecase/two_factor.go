package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/auth/domain"
)

// Labels shown in authenticator apps.
const (
	totpIssuer  = "trisse-home"
	totpAccount = "owner"
)

// checkSecondFactor verifies a TOTP or recovery code and updates the replay
// and recovery-code state of the account (the caller persists it).
func (d Deps) checkSecondFactor(a *domain.Account, code string, now time.Time) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.ErrTOTPCodeRequired
	}
	if domain.IsRecoveryCodeFormat(code) {
		return a.UseRecoveryCode(code)
	}
	step, ok := d.OTP.Verify(a.TwoFactor().Secret, code, now)
	if !ok {
		return domain.ErrInvalidTOTPCode
	}
	return a.AcceptStep(step)
}

type twoFactor struct{ deps Deps }

// NewTwoFactor returns the TwoFactor use cases.
func NewTwoFactor(deps Deps) TwoFactor { return &twoFactor{deps: deps} }

// Status implements TwoFactor.
func (uc *twoFactor) Status(ctx context.Context) (TwoFactorStatus, error) {
	a, err := uc.deps.Accounts.Get(ctx)
	if err != nil {
		return TwoFactorStatus{}, err
	}
	return TwoFactorStatus{Enabled: a.TOTPEnabled(), RecoveryCodesLeft: len(a.TwoFactor().RecoveryCodes)}, nil
}

// Start implements TwoFactor: a new secret is stored as pending until it is
// confirmed with a valid code, so a half-finished setup never locks the user out.
func (uc *twoFactor) Start(ctx context.Context) (TwoFactorEnrollment, error) {
	a, err := uc.deps.Accounts.Get(ctx)
	if err != nil {
		return TwoFactorEnrollment{}, err
	}
	secret, err := uc.deps.OTP.NewSecret()
	if err != nil {
		return TwoFactorEnrollment{}, err
	}
	if err := a.StartTOTPEnrollment(secret); err != nil {
		return TwoFactorEnrollment{}, err
	}
	if err := uc.deps.Accounts.Update(ctx, a); err != nil {
		return TwoFactorEnrollment{}, fmt.Errorf("update account: %w", err)
	}
	return TwoFactorEnrollment{Secret: secret, URI: uc.deps.OTP.ProvisioningURI(secret, totpAccount, totpIssuer)}, nil
}

// Confirm implements TwoFactor and returns the recovery codes (shown once).
func (uc *twoFactor) Confirm(ctx context.Context, code string) ([]string, error) {
	a, err := uc.deps.Accounts.Get(ctx)
	if err != nil {
		return nil, err
	}
	pending := a.TwoFactor().PendingSecret
	if pending == "" {
		return nil, domain.ErrTOTPNoEnrollment
	}
	step, ok := uc.deps.OTP.Verify(pending, code, uc.deps.Clock.Now())
	if !ok {
		return nil, domain.ErrInvalidTOTPCode
	}
	codes := make([]string, 0, domain.RecoveryCodeCount)
	for range domain.RecoveryCodeCount {
		c, err := uc.deps.OTP.NewRecoveryCode()
		if err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	if err := a.ConfirmTOTP(step, codes); err != nil {
		return nil, err
	}
	if err := uc.deps.Accounts.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}
	return codes, nil
}

// Disable implements TwoFactor; it requires the password and a current code.
func (uc *twoFactor) Disable(ctx context.Context, req DisableTwoFactorRequest) error {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(req.ClientKey, now) {
		return domain.ErrTooManyAttempts
	}
	a, err := uc.deps.Accounts.Get(ctx)
	if err != nil {
		return err
	}
	if !a.TOTPEnabled() {
		return domain.ErrTOTPNotEnabled
	}
	ok, err := uc.deps.Hasher.Verify(req.Password, a.PasswordHash())
	if err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		uc.deps.Limiter.RecordFailure(req.ClientKey, now)
		return domain.ErrInvalidCredentials
	}
	if err := uc.deps.checkSecondFactor(&a, req.Code, now); err != nil {
		if errors.Is(err, domain.ErrInvalidTOTPCode) {
			uc.deps.Limiter.RecordFailure(req.ClientKey, now)
		}
		return err
	}
	if err := a.DisableTOTP(); err != nil {
		return err
	}
	if err := uc.deps.Accounts.Update(ctx, a); err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	return nil
}
