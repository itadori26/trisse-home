package usecase

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// fakeOTP accepts "<step>" style codes: the code is the step number padded
// to six digits, which keeps tests independent of real TOTP math.
type fakeOTP struct{ codes int }

func (*fakeOTP) NewSecret() (string, error)                            { return "SECRET", nil }
func (*fakeOTP) ProvisioningURI(secret, account, issuer string) string { return "otpauth://" + secret }
func (*fakeOTP) Verify(secret, code string, now time.Time) (int64, bool) {
	step := now.Unix() / 30
	if code == fmt.Sprintf("%06d", step%1_000_000) && secret != "" {
		return step, true
	}
	return 0, false
}
func (f *fakeOTP) NewRecoveryCode() (string, error) {
	f.codes++
	return fmt.Sprintf("code-%04d", f.codes), nil
}

func currentCode(c *fakeClock) string { return fmt.Sprintf("%06d", (c.now.Unix()/30)%1_000_000) }

func TestTwoFactorLifecycleAndLogin(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.OTP = &fakeOTP{}
	tf := NewTwoFactor(e.deps)

	if _, err := tf.Confirm(ctx, "000000"); !errors.Is(err, domain.ErrTOTPNoEnrollment) {
		t.Fatalf("confirm without start: %v", err)
	}
	enr, err := tf.Start(ctx)
	if err != nil || enr.URI == "" {
		t.Fatalf("start: %v", err)
	}
	if _, err := tf.Confirm(ctx, "999999"); !errors.Is(err, domain.ErrInvalidTOTPCode) {
		t.Fatalf("wrong code confirmed: %v", err)
	}
	codes, err := tf.Confirm(ctx, currentCode(e.clock))
	if err != nil || len(codes) != domain.RecoveryCodeCount {
		t.Fatalf("confirm: %v %d", err, len(codes))
	}

	login := NewLogin(e.deps)
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("missing code must be asked for, got %v", err)
	}
	// The code used for confirmation cannot be replayed.
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, TOTPCode: currentCode(e.clock), ClientKey: "ip"}); !errors.Is(err, domain.ErrInvalidTOTPCode) {
		t.Fatalf("replay accepted: %v", err)
	}
	e.clock.now = e.clock.now.Add(30 * time.Second)
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, TOTPCode: currentCode(e.clock), ClientKey: "ip"}); err != nil {
		t.Fatalf("login with code: %v", err)
	}
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, TOTPCode: codes[0], ClientKey: "ip2"}); err != nil {
		t.Fatalf("login with recovery code: %v", err)
	}
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, TOTPCode: codes[0], ClientKey: "ip2"}); !errors.Is(err, domain.ErrInvalidTOTPCode) {
		t.Fatalf("recovery code reused: %v", err)
	}
	st, _ := tf.Status(ctx)
	if !st.Enabled || st.RecoveryCodesLeft != domain.RecoveryCodeCount-1 {
		t.Fatalf("status %+v", st)
	}
	if err := tf.Disable(ctx, DisableTwoFactorRequest{Password: "wrong", Code: codes[1], ClientKey: "ip3"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("disable with wrong password: %v", err)
	}
	if err := tf.Disable(ctx, DisableTwoFactorRequest{Password: goodPassword, Code: codes[1], ClientKey: "ip3"}); err != nil {
		t.Fatal(err)
	}
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip4"}); err != nil {
		t.Fatalf("login without 2FA after disable: %v", err)
	}
}
