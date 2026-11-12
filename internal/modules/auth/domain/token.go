package domain

import (
	"regexp"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// APITokenPrefix marks personal API tokens, so they are recognisable in
// configuration files and by secret scanners.
const APITokenPrefix = "th_"

// MaxAPITokens bounds the number of tokens.
const MaxAPITokens = 50

// tokenUseRecordInterval limits how often last-use data is written.
const tokenUseRecordInterval = time.Hour

var scopePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,30}:(read|write)$`)

// APIToken is a personal access token of the owner with scopes such as
// "todos:write". Only its hash is stored.
type APIToken struct {
	ID         string
	Name       string
	Hash       TokenHash
	Scopes     []string
	CreatedAt  time.Time
	ExpiresAt  time.Time // zero: never
	LastUsedAt time.Time
	LastUsedIP string
}

// NewAPIToken validates and creates a token.
func NewAPIToken(id, name string, hash TokenHash, scopes []string, now, expiresAt time.Time) (APIToken, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 60 {
		return APIToken{}, kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	clean := []string{}
	for _, s := range scopes {
		s = strings.ToLower(strings.TrimSpace(s))
		if !scopePattern.MatchString(s) || strings.HasPrefix(s, "auth:") {
			return APIToken{}, kernel.NewValidationErrorf("scopes", "Invalid scope {scope}: use module:read or module:write.", "scope", s)
		}
		if !slices.Contains(clean, s) {
			clean = append(clean, s)
		}
	}
	if len(clean) == 0 {
		return APIToken{}, kernel.NewValidationError("scopes", "Choose at least one scope.")
	}
	if !expiresAt.IsZero() && !expiresAt.After(now) {
		return APIToken{}, kernel.NewValidationError("expiresAt", "The expiry must be in the future.")
	}
	slices.Sort(clean)
	return APIToken{ID: id, Name: name, Hash: hash, Scopes: clean, CreatedAt: now, ExpiresAt: expiresAt}, nil
}

// Expired reports whether the token is no longer valid at now.
func (t APIToken) Expired(now time.Time) bool {
	return !t.ExpiresAt.IsZero() && !now.Before(t.ExpiresAt)
}

// RecordUse notes the use and reports whether it should be persisted (at
// most once per hour, so tokens used by automations do not write on every
// request).
func (t *APIToken) RecordUse(now time.Time, ip string) bool {
	if now.Sub(t.LastUsedAt) < tokenUseRecordInterval && ip == t.LastUsedIP {
		return false
	}
	t.LastUsedAt, t.LastUsedIP = now, ip
	return true
}

// TokenMayAccess decides whether scopes allow method on path. A module's
// write scope includes reading. Auth endpoints are never reachable with a
// token, so a leaked token cannot create more tokens or change the account.
func TokenMayAccess(scopes []string, method, path string) bool {
	if !strings.HasPrefix(path, "/api/") {
		return false
	}
	module, _, _ := strings.Cut(strings.TrimPrefix(path, "/api/"), "/")
	if module == "" || module == "auth" {
		return false
	}
	if slices.Contains(scopes, module+":write") {
		return true
	}
	return (method == "GET" || method == "HEAD") && slices.Contains(scopes, module+":read")
}
