package kernel

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"testing"
)

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func TestTargetFailure(t *testing.T) {
	wrap := func(err error) error { return &url.Error{Op: "Get", URL: "https://10.0.0.7:8006/api2/json", Err: err} }
	cases := map[string]struct {
		err  error
		want string
	}{
		"blocked":     {wrap(ErrTargetBlocked), "Connections to this address are not allowed."},
		"auth":        {fmt.Errorf("%w: token rejected (401)", ErrTargetAuth), "The server rejected the login. Check user name, password or token."},
		"not found":   {fmt.Errorf("%w: folder", ErrTargetNotFound), "The server answered, but the path was not found. Check the address."},
		"certificate": {wrap(x509.UnknownAuthorityError{}), "The certificate of the server is not trusted."},
		"timeout":     {wrap(timeoutErr{}), "The server did not answer in time."},
		"deadline":    {fmt.Errorf("get: %w", context.DeadlineExceeded), "The server did not answer in time."},
		"refused":     {wrap(&net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}), "The server could not be reached. Check the address and port."},
		"dns":         {wrap(&net.DNSError{Err: "no such host", Name: "nas.lan"}), "The server could not be reached. Check the address and port."},
		"other":       {errors.New("unexpected answer: 500 Internal Server Error"), "The server sent an unexpected answer."},
	}
	for name, c := range cases {
		e := TargetFailure(ErrUnavailable, c.err)
		if e.Message != c.want || !errors.Is(e, ErrUnavailable) || !errors.Is(e.Cause, c.err) {
			t.Errorf("%s: %q", name, e.Message)
		}
		// No host, port or raw error text towards the client.
		for _, leak := range []string{"10.0.0.7", "8006", "nas.lan", "refused", "500"} {
			if strings.Contains(e.Error(), leak) {
				t.Errorf("%s: message leaks %q", name, leak)
			}
		}
	}
}
