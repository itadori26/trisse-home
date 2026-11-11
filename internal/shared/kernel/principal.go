package kernel

import (
	"context"
	"slices"
)

// Principal is the authenticated account of a request. The auth module puts
// it into the request context; other modules read it, e.g. to show a guest
// only the tabs shared with it.
type Principal struct {
	AccountID string
	// Role is "owner" or "guest".
	Role string
	// Tabs are the dashboard tabs a guest may see (empty for the owner).
	Tabs []string
}

// IsGuest reports whether the principal is a read-only guest.
func (p Principal) IsGuest() bool { return p.Role == "guest" }

// CanSeeTab reports whether the principal may see a dashboard tab.
func (p Principal) CanSeeTab(id string) bool { return !p.IsGuest() || slices.Contains(p.Tabs, id) }

type principalKey struct{}

// WithPrincipal returns ctx carrying p.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom returns the principal of ctx; requests without one (jobs,
// tests) act as the owner.
func PrincipalFrom(ctx context.Context) Principal {
	if p, ok := ctx.Value(principalKey{}).(Principal); ok {
		return p
	}
	return Principal{AccountID: "owner", Role: "owner"}
}
