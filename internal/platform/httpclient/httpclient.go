package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

// UserAgent identifies this application to external services.
const UserAgent = "trisse-home (+https://github.com/topics/self-hosted)"

// Option configures a client.
type Option func(*options)

type options struct {
	timeout            time.Duration
	insecureSkipVerify bool
}

// WithTimeout sets the overall request timeout.
func WithTimeout(d time.Duration) Option { return func(o *options) { o.timeout = d } }

// WithInsecureTLS disables certificate verification. Only meant for status
// checks of LAN services with self-signed certificates.
func WithInsecureTLS() Option { return func(o *options) { o.insecureSkipVerify = true } }

const defaultTimeout = 10 * time.Second

// New returns an *http.Client that sets a User-Agent on every request.
func New(opts ...Option) *http.Client {
	o := options{timeout: defaultTimeout}
	for _, opt := range opts {
		opt(&o)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if o.insecureSkipVerify {
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
