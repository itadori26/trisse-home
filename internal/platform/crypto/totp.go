package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // RFC 6238 TOTP as implemented by all authenticator apps uses HMAC-SHA1
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TOTP parameters compatible with common authenticator apps.
const (
	totpPeriod      = 30
	totpDigits      = 6
	totpSecretBytes = 20
	// totpSkew accepts the previous and next code to tolerate clock drift.
	totpSkew          = 1
	recoveryCodeBytes = 5
)

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// TOTP implements time-based one-time passwords (RFC 6238).
type TOTP struct{}

// NewSecret returns a random base32 secret.
func (TOTP) NewSecret() (string, error) {
	b := make([]byte, totpSecretBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate totp secret: %w", err)
	}
	return b32.EncodeToString(b), nil
}

// ProvisioningURI returns the otpauth:// URI encoded in the setup QR code.
func (TOTP) ProvisioningURI(secret, account, issuer string) string {
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", fmt.Sprint(totpDigits))
	q.Set("period", fmt.Sprint(totpPeriod))
	label := url.PathEscape(issuer + ":" + account)
	return "otpauth://totp/" + label + "?" + q.Encode()
}

// Verify checks code against secret at now. It returns the matching time
// step so callers can reject reuse of the same code.
func (TOTP) Verify(secret, code string, now time.Time) (int64, bool) {
	code = strings.ReplaceAll(strings.TrimSpace(code), " ", "")
	if len(code) != totpDigits {
		return 0, false
	}
	key, err := b32.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return 0, false
	}
	step := now.Unix() / totpPeriod
	for delta := int64(-totpSkew); delta <= totpSkew; delta++ {
		if subtle.ConstantTimeCompare([]byte(hotp(key, step+delta)), []byte(code)) == 1 {
			return step + delta, true
		}
	}
	return 0, false
}

// Code returns the code for now (used in tests and for diagnostics).
func (TOTP) Code(secret string, now time.Time) (string, error) {
	key, err := b32.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}
	return hotp(key, now.Unix()/totpPeriod), nil
}

// NewRecoveryCode returns a random code like "k3x9-q2m7".
func (TOTP) NewRecoveryCode() (string, error) {
	b := make([]byte, recoveryCodeBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate recovery code: %w", err)
	}
	s := strings.ToLower(b32.EncodeToString(b))
	return s[:4] + "-" + s[4:], nil
}

func hotp(key []byte, counter int64) string {
	mac := hmac.New(sha1.New, key)
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(counter)) //nolint:gosec // counter is a positive unix time step
	mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%0*d", totpDigits, value%1_000_000)
}
