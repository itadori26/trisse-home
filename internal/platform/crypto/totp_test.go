package crypto

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"
)

// RFC 6238 Appendix B test vectors (SHA1, 8 digits) truncated to 6 digits.
func TestTOTPRFCVectors(t *testing.T) {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("12345678901234567890"))
	vectors := map[int64]string{59: "287082", 1111111109: "081804", 1234567890: "005924", 2000000000: "279037"}
	for unix, want := range vectors {
		got, err := TOTP{}.Code(secret, time.Unix(unix, 0))
		if err != nil || got != want {
			t.Errorf("t=%d: got %s, want %s", unix, got, want)
		}
	}
}

func TestTOTPVerifyWithSkew(t *testing.T) {
	var otp TOTP
	secret, _ := otp.NewSecret()
	now := time.Unix(1_800_000_000, 0)
	prev, _ := otp.Code(secret, now.Add(-30*time.Second))
	if step, ok := otp.Verify(secret, prev, now); !ok || step != now.Unix()/30-1 {
		t.Fatalf("previous code rejected: %v %d", ok, step)
	}
	old, _ := otp.Code(secret, now.Add(-90*time.Second))
	if _, ok := otp.Verify(secret, old, now); ok {
		t.Fatal("code outside the window accepted")
	}
	if _, ok := otp.Verify(secret, "12345", now); ok {
		t.Fatal("short code accepted")
	}
}

func TestProvisioningURIAndRecoveryCode(t *testing.T) {
	var otp TOTP
	uri := otp.ProvisioningURI("ABC", "owner", "trisse-home")
	if !strings.HasPrefix(uri, "otpauth://totp/trisse-home:owner?") || !strings.Contains(uri, "secret=ABC") {
		t.Fatalf("unexpected uri %s", uri)
	}
	code, err := otp.NewRecoveryCode()
	if err != nil || len(code) != 9 || code[4] != '-' {
		t.Fatalf("unexpected recovery code %q", code)
	}
}
