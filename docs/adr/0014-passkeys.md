# ADR 0014: Passkeys as second factor with a minimal WebAuthn implementation

- Status: Accepted
- Date: 2026-11-12
- Extends: [ADR 0009](0009-two-factor-authentication.md)

## Context

TOTP codes (ADR 0009) work everywhere but have to be typed. Passkeys
(WebAuthn) are faster and phishing-resistant: the browser binds the
signature to the origin (#39).

## Decision

- Passkeys are an additional **second factor**: after the correct password
  the login answers `202` with a single-use ticket and a challenge; the
  browser signs it with one of the account's passkeys
  (`POST /api/auth/passkeys/login`). TOTP and recovery codes keep working
  as alternatives. An account whose only second factor is a passkey gets
  recovery codes with its first passkey. Without TOTP set up, six-digit
  codes are never accepted (only recovery codes).
- Owner and guests can register up to 10 passkeys each (Settings → Account).
- **Own minimal implementation** in `internal/platform/webauthn` (standard
  library only, ~400 lines incl. a bounded CBOR decoder) instead of
  `go-webauthn/webauthn`, whose dependency tree (CBOR, TPM, JWT libraries)
  conflicts with ADR 0011. Scope: attestation `none` is requested and
  attestation statements are **not** evaluated (no enterprise attestation
  needed for a home start page); algorithms ES256, EdDSA and RS256; checks
  of ceremony type, challenge, origin, RP ID hash, user presence and the
  signature; signature counters must increase (clone detection), counters
  of 0 are accepted for authenticators without counters.
- WebAuthn needs a fixed origin and a secure context: passkeys are only
  offered when `PUBLIC_URL` is set to an `https://` origin (or
  `http://localhost`). Behind a plain `http://<ip>:8090` they stay
  disabled with an explanation in the settings.
- The browser API is isolated in `web/src/features/auth/adapter/browser`;
  use cases and UI stay testable without `navigator.credentials`.

## Consequences

- No new dependency; the implementation is covered by tests with a
  simulated authenticator (accepted ceremonies for ES256/EdDSA, rejected
  wrong challenge, phishing origin, foreign RP ID, missing user presence,
  tampered data, wrong ceremony type, foreign key, CBOR bombs) and an
  end-to-end test over HTTP.
- Passwordless login (passkey without password) is a possible follow-up;
  the ticket design keeps the password as first factor for now.
