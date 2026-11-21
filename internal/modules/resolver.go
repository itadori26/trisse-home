package modules

import (
	"context"
	"errors"
	"sync"
)

// ResolvedConnection is the name, address and credential of a named
// connection of the integrations module (#610).
type ResolvedConnection struct {
	Name       string
	URL        string
	Credential map[string]string
}

// ErrNoConnection means a named connection does not exist, belongs to
// another service or cannot be read; the module then uses its own settings.
var ErrNoConnection = errors.New("no usable connection")

// ConnectionResolver looks up the named connection a widget chose
// ("connectionId"). It is implemented by the integrations module.
type ConnectionResolver interface {
	ResolveConnection(ctx context.Context, service, id string) (ResolvedConnection, error)
}

// LateResolver forwards to a resolver set after the modules are built: the
// integrations module needs the other modules' specs, so it is built after
// the modules that read through it. Until Set, and when the integrations
// module is switched off, every lookup answers ErrNoConnection.
type LateResolver struct {
	mu sync.RWMutex
	r  ConnectionResolver
}

// Set installs the resolver.
func (l *LateResolver) Set(r ConnectionResolver) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.r = r
}

// ResolveConnection implements ConnectionResolver.
func (l *LateResolver) ResolveConnection(ctx context.Context, service, id string) (ResolvedConnection, error) {
	l.mu.RLock()
	r := l.r
	l.mu.RUnlock()
	if r == nil || id == "" {
		return ResolvedConnection{}, ErrNoConnection
	}
	return r.ResolveConnection(ctx, service, id)
}
