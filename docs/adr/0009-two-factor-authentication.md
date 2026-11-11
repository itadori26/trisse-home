# 0009 – Two-factor authentication with TOTP

- Status: Accepted
- Date: 2026-11-11

## Context

The start page shows personal data and server details; a leaked or guessed
password should not be enough to log in (issue #30).

## Decision

- Optional TOTP (RFC 6238, HMAC-SHA1, 6 digits, 30 s, ±1 step) because every
  authenticator app supports it and it needs no external service. The
  algorithm is implemented in `platform/crypto` (no dependency) and verified
  against the RFC test vectors; use cases depend on a `OneTimePasswords` port.
- Enrollment is two-step: a pending secret is stored and only activated after
  a valid code was entered, so a half-finished setup never locks the user out.
- Replay protection: the time step of the last accepted code is stored; codes
  of the same or an earlier step are rejected.
- Ten recovery codes are shown once, stored as SHA-256 hashes and consumed on
  use. The secret is encrypted with `SECRET_KEY` (AES-256-GCM).
- Login answers with 422 (`field: totpCode`) when the second factor is
  missing; wrong codes count towards the existing rate limit.
- Disabling requires the password and a current or recovery code.
  `reset-account` removes the second factor together with the account.

## Consequences

- Passkeys (WebAuthn) are a separate follow-up; they need a stable origin
  (HTTPS domain) which the default IP:port deployment does not have.
- Losing both the authenticator and all recovery codes requires
  `reset-account` on the server.
