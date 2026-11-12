package httpclient

import (
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

// UserAgent identifies this application to external services.
const UserAgent = "trisse-home (+https://github.com/topics/self-hosted)"

// Option configures a client.
type Option func(*options)

type options struct {
	timeout            time.Duration
	insecureSkipVerify bool
	pinnedSHA256       []byte
}

// WithTimeout sets the overall request timeout.
func WithTimeout(d time.Duration) Option { return func(o *options) { o.timeout = d } }

// WithInsecureTLS disables certificate verification. Only meant for status
// checks of LAN services with self-signed certificates.
func WithInsecureTLS() Option { return func(o *options) { o.insecureSkipVerify = true } }

// WithPinnedCertificate accepts exactly the server certificate whose
// SHA-256 fingerprint (hex, colons allowed) matches, instead of verifying it
// against the system CAs. Meant for appliances with a self-signed
// certificate such as Proxmox VE: safer than WithInsecureTLS because any
// other certificate is rejected. An invalid fingerprint rejects everything.
func WithPinnedCertificate(fingerprint string) Option {
	return func(o *options) {
		fp, err := hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(fingerprint), ":", ""))
		if err != nil || len(fp) != sha256.Size {
			fp = make([]byte, sha256.Size) // matches no certificate
		}
		o.pinnedSHA256 = fp
	}
}

// errFingerprint is returned when the server certificate is not the pinned one.
var errFingerprint = errors.New("the server certificate does not match the pinned fingerprint")

const defaultTimeout = 10 * time.Second

// New returns an *http.Client that sets a User-Agent on every request.
func New(opts ...Option) *http.Client {
	o := options{timeout: defaultTimeout}
	for _, opt := range opts {
		opt(&o)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	switch {
	case o.pinnedSHA256 != nil:
		pin := o.pinnedSHA256
		transport.TLSClientConfig = &tls.Config{
			// The chain is not verified against CAs; VerifyConnection checks
			// the pinned fingerprint instead, so only that certificate passes.
			InsecureSkipVerify: true, //nolint:gosec // replaced by certificate pinning below
			VerifyConnection: func(cs tls.ConnectionState) error {
				if len(cs.PeerCertificates) == 0 {
					return errFingerprint
				}
				sum := sha256.Sum256(cs.PeerCertificates[0].Raw)
				if subtle.ConstantTimeCompare(sum[:], pin) != 1 {
					return errFingerprint
				}
				return nil
			},
		}
	case o.insecureSkipVerify:
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // explicit opt-in for LAN health checks
	}
	return &http.Client{Timeout: o.timeout, Transport: userAgentTransport{next: transport}}
}

type userAgentTransport struct{ next http.RoundTripper }

func (t userAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		r = r.Clone(r.Context())
		r.Header.Set("User-Agent", UserAgent)
	}
	return t.next.RoundTrip(r)
}
