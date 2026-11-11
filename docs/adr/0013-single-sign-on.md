# ADR 0013: Single sign-on with OpenID Connect and trusted proxy headers

- Status: Accepted
- Date: 2026-11-11
- Supersedes: the "no OAuth" consequence of [ADR 0005](0005-single-user-auth.md)

## Context

Many home labs already run an identity provider (Authelia, Authentik,
Keycloak, Pocket ID) with one login and one second factor for all
services (#79). A separate password for the start page is friction, and
users ask to manage access centrally.

## Decision

Two optional ways to prove the **owner's** identity; both end in the
normal session creation, so cookies, CSRF protection, the guard and the
session overview stay unchanged:

1. **OpenID Connect** (authorization code flow with PKCE S256, `state`
   bound to the browser in a short-lived cookie, `nonce` in the ID token).
   Only subjects or *verified* e-mail addresses listed in
   `OIDC_ALLOWED_SUBJECTS` / `OIDC_ALLOWED_EMAILS` are accepted.
2. **Trusted proxy header** (forward auth): the user name in
   `AUTH_PROXY_HEADER` (e.g. `Remote-User`) is believed **only** when the
   TCP peer is in `TRUSTED_PROXIES`; from every other source it is ignored.
   Only names in `AUTH_PROXY_USERS` sign in.

The local second factor is skipped for SSO by default (the identity
provider is responsible); `SSO_REQUIRE_2FA=true` asks for it anyway.
`PASSWORD_LOGIN=false` disables the owner's password login once SSO works;
`trisse-home reset-account` still recovers. Guests keep their passwords.

**Own implementation instead of a library.** `internal/platform/oidc`
implements discovery, PKCE, the token request and ID token validation with
the standard library (~350 lines): RS256 and ES256 only (no `none`, no HMAC
with public keys), issuer, audience, expiry, issue time, nonce, JWKS cache
with refresh for key rotation. The common libraries (`coreos/go-oidc` +
`golang.org/x/oauth2` + `go-jose`) would add three dependencies for the
same checks (ADR 0011). The validation is covered by tests against an
in-process provider, including tampered claims, `alg: none`, wrong
audience/nonce/issuer and expired tokens.

## Consequences

- No new runtime dependency; the binary stays within its budget.
- We maintain security-relevant code ourselves: the tests above are the
  safety net, and the scope is deliberately small (no implicit flow, no
  refresh tokens, no userinfo endpoint, no logout propagation).
- Header auth is only safe behind a proxy that strips the header from
  clients; `docs/reverse-proxy.md` shows Authelia and Authentik setups.
