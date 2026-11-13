package kernel

import (
	"net/netip"
	"net/url"
	"strings"
)

// Owner-configured integration targets (#191).
//
// Integrations such as AdGuard, Nextcloud or Proxmox call an address the
// owner enters in the settings. ParseTarget is the one place that decides
// which of these addresses are acceptable; every module uses it in its
// settings validation instead of its own checks. The HTTP client of the
// integrations repeats the address rule for every connection
// (httpclient.WithIntegrationGuard), so a hostname that later resolves to a
// forbidden address is refused as well.
//
// Loopback and private ranges stay allowed on purpose: home-lab services
// live there. Link-local addresses (incl. the cloud metadata service at
// 169.254.169.254), multicast and unspecified addresses are never needed by
// an integration and are refused.

// TargetOption relaxes ParseTarget for a module.
type TargetOption func(*targetRules)

type targetRules struct{ path, query bool }

// WithTargetPath allows a path, e.g. a WebDAV folder or a CalDAV collection.
// Without it only the server address (scheme, host, port) is accepted.
func WithTargetPath() TargetOption { return func(r *targetRules) { r.path = true } }

// WithTargetQuery additionally allows a query, which some webhook services
// use for their token. A fragment is never allowed.
func WithTargetQuery() TargetOption { return func(r *targetRules) { r.path, r.query = true, true } }

// ParseTarget validates an owner-configured target URL: scheme http or
// https, a host, no credentials, no query or fragment, no path unless
// allowed, and no forbidden address literal. It returns the URL without a
// trailing slash. Errors are validation errors for field.
func ParseTarget(field, raw string, opts ...TargetOption) (*url.URL, error) {
	var rules targetRules
	for _, o := range opts {
		o(&rules)
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.Opaque != "" {
		return nil, NewValidationError(field, "Enter an http(s) address with a host, e.g. https://nas.local:8443.")
	}
	if u.User != nil {
		return nil, NewValidationError(field, "Remove the user name and password from the address and use the fields below.")
	}
	if u.Fragment != "" || (!rules.query && (u.RawQuery != "" || u.ForceQuery)) {
		return nil, NewValidationError(field, "The address must not contain a query (?…) or fragment (#…).")
	}
	if !rules.path && u.Path != "" && u.Path != "/" {
		return nil, NewValidationError(field, "Enter only the server address, without a path.")
	}
	if ip, err := netip.ParseAddr(strings.Trim(u.Hostname(), "[]")); err == nil && !TargetAddrAllowed(ip) {
		return nil, NewValidationError(field, "Link-local and cloud metadata addresses are not allowed.")
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.RawPath = ""
	return u, nil
}

// TargetAddrAllowed reports whether an integration may connect to ip:
// everything except link-local (169.254.0.0/16, fe80::/10, incl. the cloud
// metadata service), multicast, unspecified and 0.0.0.0/8 addresses.
func TargetAddrAllowed(ip netip.Addr) bool {
	ip = ip.Unmap()
	switch {
	case !ip.IsValid(), ip.IsUnspecified(), ip.IsLinkLocalUnicast(), ip.IsLinkLocalMulticast(),
		ip.IsInterfaceLocalMulticast(), ip.IsMulticast():
		return false
	case ip.Is4() && ip.As4()[0] == 0: // "this network"
		return false
	}
	return true
}

// TargetChanged reports whether the identity of a target changed, e.g. its
// URL, token ID or certificate fingerprint. A stored secret must then be
// entered again instead of being sent silently to the new target (#191).
func TargetChanged(old, updated []string) bool {
	if len(old) != len(updated) {
		return true
	}
	for i := range old {
		if !strings.EqualFold(strings.TrimRight(strings.TrimSpace(old[i]), "/"), strings.TrimRight(strings.TrimSpace(updated[i]), "/")) {
			return true
		}
	}
	return false
}

// SecretRequired is returned for field when a target changed but no new
// secret was entered.
func SecretRequired(field string) error {
	return NewValidationError(field, "The address or account changed: enter the password or token again.")
}
