# 0005 – Single-user auth with setup token and sessions

- Status: Accepted
- Date: 2026-11-10

## Context

The app is used by one person and reachable only from trusted networks, but
it still exposes personal data and server details and must not be claimable
by anyone else in the network after the first start.

## Decision

- First-run setup in the browser, authorised by a one-time setup token that
  is printed to the server log (or fixed via `SETUP_TOKEN`).
- Passwords are hashed with Argon2id (m=64 MiB, t=3, p=2).
- Sessions are random 256-bit tokens in an `HttpOnly`, `SameSite=Strict`
  cookie (`Secure` + `__Host-` prefix when `COOKIE_SECURE=true`). Only the
  SHA-256 of the token is stored. Sessions slide and expire after
  `SESSION_TTL`.
- CSRF protection uses a per-session synchronizer token sent in
  `X-CSRF-Token`, plus Go's `http.CrossOriginProtection` (Sec-Fetch-Site /
  Origin) for all state-changing requests, including login.
- Failed logins are rate-limited per client IP (5 per 15 minutes).
- An IP allowlist runs before everything else.
- `trisse-home reset-account` removes the account for recovery.

## Consequences

- No user management, roles or OAuth – deliberately out of scope.
- Integration credentials are encrypted with AES-256-GCM using `SECRET_KEY`.
