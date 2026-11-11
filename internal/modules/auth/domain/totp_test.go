package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTOTPLifecycle(t *testing.T) {
	a, _ := NewAccount("hash", time.Now())
	if err := a.ConfirmTOTP(1, nil); !errors.Is(err, ErrTOTPNoEnrollment) {
		t.Fatalf("confirm without enrollment: %v", err)
	}
	_ = a.StartTOTPEnrollment("SECRET")
	if a.TOTPEnabled() {
		t.Fatal("pending secret must not enable 2FA")
	}
	if err := a.ConfirmTOTP(100, []string{"abcd-efgh", "ijkl-mnop"}); err != nil || !a.TOTPEnabled() {
		t.Fatalf("confirm: %v", err)
	}
	if err := a.StartTOTPEnrollment("OTHER"); !errors.Is(err, ErrTOTPAlreadyEnabled) {
		t.Fatalf("re-enroll: %v", err)
	}
	if err := a.AcceptStep(100); !errors.Is(err, ErrInvalidTOTPCode) {
		t.Fatal("replayed step accepted")
	}
	if err := a.AcceptStep(101); err != nil {
		t.Fatal(err)
	}
	if err := a.UseRecoveryCode(" ABCD EFGH "); err != nil {
		t.Fatalf("recovery code with other formatting: %v", err)
	}
	if err := a.UseRecoveryCode("abcd-efgh"); !errors.Is(err, ErrInvalidTOTPCode) {
		t.Fatal("recovery code reused")
	}
	if len(a.TwoFactor().RecoveryCodes) != 1 {
		t.Fatal("recovery code not consumed")
	}
	if !IsRecoveryCodeFormat("ijkl-mnop") || IsRecoveryCodeFormat("123456") {
		t.Fatal("format detection wrong")
	}
	if err := a.DisableTOTP(); err != nil || a.TOTPEnabled() {
		t.Fatal("disable failed")
	}
	if err := a.DisableTOTP(); !errors.Is(err, ErrTOTPNotEnabled) {
		t.Fatal("double disable")
	}
}
