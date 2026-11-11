package domain

import (
	"slices"
	"strings"
)

// ExternalIdentity is a user proven by an identity provider or by a
// trusted reverse proxy.
type ExternalIdentity struct {
	Provider      string // "oidc" or "header"
	Subject       string
	Email         string
	EmailVerified bool
}

// IdentityMapping lists who may sign in as the owner via SSO.
type IdentityMapping struct {
	Subjects []string
	Emails   []string
}

// Allows reports whether id may sign in as the owner: its subject is
// listed, or its verified e-mail address is (case-insensitive).
func (m IdentityMapping) Allows(id ExternalIdentity) bool {
	if id.Subject != "" && slices.Contains(m.Subjects, id.Subject) {
		return true
	}
	if !id.EmailVerified || id.Email == "" {
		return false
	}
	return slices.ContainsFunc(m.Emails, func(e string) bool { return strings.EqualFold(e, id.Email) })
}
