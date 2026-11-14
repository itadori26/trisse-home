package httpclient

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"syscall"
	"time"

	"trisse-home/internal/shared/kernel"
)

// UserAgent identifies this application to external services.
const UserAgent = "trisse-home (+https://github.com/topics/self-hosted)"

// Option configures a client.
type Option func(*options)

type options struct {
	timeout            time.Duration
	insecureSkipVerify bool
	pinnedSHA256       []byte
	guard              bool
	integration        bool
	privateOnly        bool
	maxResponseBytes   int64
}

// WithIntegrationGuard hardens a client for owner-configured integrations
// such as AdGuard, Nextcloud or Proxmox (#191): every connection is checked
// against kernel.TargetAddrAllowed on the resolved address (so DNS cannot
// get around the check of kernel.ParseTarget), redirects are not followed,
// no proxy is used and responses are limited to DefaultMaxResponseBytes
// unless WithMaxResponseBytes says otherwise. Loopback and LAN addresses
// stay allowed: home-lab services live there.
func WithIntegrationGuard() Option { return func(o *options) { o.integration = true } }

// WithPrivateOnly narrows WithIntegrationGuard to loopback and private
// addresses (RFC 1918, RFC 4193, 100.64.0.0/10), checked on the resolved
// address of every connection. Used where the owner did not explicitly allow
// public addresses, e.g. the generic JSON widget (#48).
func WithPrivateOnly() Option {
	return func(o *options) { o.integration, o.privateOnly = true, true }
}

// PrivateAddr reports whether ip is loopback or in a private range.
func PrivateAddr(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsLoopback() || ip.IsPrivate() || cgnat.Contains(ip)
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

func privateControl(network, address string, c syscall.RawConn) error {
	if err := integrationControl(network, address, c); err != nil {
		return err
	}
	host, _, _ := net.SplitHostPort(address) // checked by integrationControl
	if ip, err := netip.ParseAddr(host); err != nil || !PrivateAddr(ip) {
		return ErrBlockedAddress
	}
	return nil
}

// WithMaxResponseBytes limits the size of a response body; reading more
// returns ErrResponseTooLarge.
func WithMaxResponseBytes(n int64) Option { return func(o *options) { o.maxResponseBytes = n } }

// DefaultMaxResponseBytes is the response limit of integration clients.
const DefaultMaxResponseBytes = 32 << 20

// ErrResponseTooLarge is returned when a response exceeds the limit.
var ErrResponseTooLarge = errors.New("the response is too large")

func integrationControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip, err := netip.ParseAddr(host)
	if err != nil || !kernel.TargetAddrAllowed(ip) {
		return ErrBlockedAddress
	}
	return nil
}

// WithInternalAddressGuard refuses connections to loopback, link-local
// (incl. cloud metadata at 169.254.169.254), unspecified and multicast
// addresses and limits redirects to five. The check runs on the resolved
// address of every connection, so DNS rebinding and redirects cannot get
// around it. LAN addresses stay allowed: bookmarks of a home server point
// to LAN devices on purpose.
func WithInternalAddressGuard() Option { return func(o *options) { o.guard = true } }

// ErrBlockedAddress is returned when the guard refuses a connection.
var ErrBlockedAddress = kernel.ErrTargetBlocked

// Blocked reports whether the guard refuses ip.
func Blocked(ip netip.Addr) bool {
	ip = ip.Unmap()
	return !ip.IsValid() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() ||
		(ip.Is4() && ip.As4()[0] == 0) // 0.0.0.0/8 "this network"
}

func guardControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip, err := netip.ParseAddr(host)
	if err != nil || Blocked(ip) {
		return ErrBlockedAddress
	}
	return nil
}

const maxRedirects = 5

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
	if o.integration && o.maxResponseBytes == 0 {
		o.maxResponseBytes = DefaultMaxResponseBytes
	}
	var rt http.RoundTripper = userAgentTransport{next: transport}
	if o.maxResponseBytes > 0 {
		rt = limitTransport{next: rt, max: o.maxResponseBytes}
	}
	client := &http.Client{Timeout: o.timeout, Transport: rt}
	if o.integration {
		control := integrationControl
		if o.privateOnly {
			control = privateControl
		}
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second, Control: control}
		transport.DialContext = dialer.DialContext
		transport.Proxy = nil // a proxy would connect on our behalf and bypass the check
		// Integrations talk to one configured server; a redirect elsewhere is
		// answered with the redirect response itself instead of following it.
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	}
	if o.guard {
		dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second, Control: guardControl}
		transport.DialContext = dialer.DialContext
		transport.Proxy = nil // a proxy would connect on our behalf and bypass the check
		client.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		}
	}
	return client
}

type userAgentTransport struct{ next http.RoundTripper }

func (t userAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		r = r.Clone(r.Context())
		r.Header.Set("User-Agent", UserAgent)
	}
	return t.next.RoundTrip(r)
}

// limitTransport limits the size of response bodies.
type limitTransport struct {
	next http.RoundTripper
	max  int64
}

func (t limitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	res, err := t.next.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if res.ContentLength > t.max {
		_ = res.Body.Close()
		return nil, fmt.Errorf("%w (%d bytes)", ErrResponseTooLarge, res.ContentLength)
	}
	res.Body = &limitedBody{rc: res.Body, left: t.max}
	return res, nil
}

type limitedBody struct {
	rc   io.ReadCloser
	left int64
}

func (b *limitedBody) Read(p []byte) (int, error) {
	if b.left <= 0 {
		// One more byte tells "exactly at the limit" from "too large".
		var one [1]byte
		if n, _ := b.rc.Read(one[:]); n > 0 {
			return 0, ErrResponseTooLarge
		}
		return 0, io.EOF
	}
	if int64(len(p)) > b.left {
		p = p[:b.left]
	}
	n, err := b.rc.Read(p)
	b.left -= int64(n)
	return n, err
}

func (b *limitedBody) Close() error { return b.rc.Close() }

// IntegrationDialContext dials TCP connections to owner-configured
// integrations that do not speak HTTP, such as IMAP (#166), with the same
// address check as WithIntegrationGuard: link-local, metadata, multicast and
// unspecified addresses are refused on the resolved address.
func IntegrationDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{Timeout: 10 * time.Second, Control: integrationControl}
	return d.DialContext(ctx, network, address)
}
