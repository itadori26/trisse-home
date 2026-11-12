package usecase

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
)

func TestConfirmIdentity(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.OTP = &fakeOTP{}
	confirm := NewConfirmIdentity(e.deps)

	if err := confirm.Confirm(ctx, "wrong", "", "ip"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	if err := confirm.Confirm(ctx, goodPassword, "", "ip"); err != nil {
		t.Fatalf("password only without 2FA: %v", err)
	}

	tf := NewTwoFactor(e.deps)
	_, _ = tf.Start(ctx)
	if _, err := tf.Confirm(ctx, currentCode(e.clock)); err != nil {
		t.Fatal(err)
	}
	e.clock.now = e.clock.now.Add(30 * time.Second)
	if err := confirm.Confirm(ctx, goodPassword, "", "ip"); err == nil {
		t.Fatal("2FA code must be required once enabled")
	}
	if err := confirm.Confirm(ctx, goodPassword, currentCode(e.clock), "ip"); err != nil {
		t.Fatalf("password and code: %v", err)
	}
	if err := confirm.Confirm(ctx, goodPassword, currentCode(e.clock), "ip"); !errors.Is(err, domain.ErrInvalidTOTPCode) {
		t.Fatalf("replayed code: %v", err)
	}
}
