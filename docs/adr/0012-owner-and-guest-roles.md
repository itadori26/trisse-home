# ADR 0012: Owner and read-only guest accounts

- Status: Accepted
- Date: 2026-11-11
- Supersedes: the single-user scope of [ADR 0005](0005-single-user-auth.md)

## Context

ADR 0005 chose exactly one account: no user management, no roles, nothing
to misconfigure. Households want to show a tab (shopping list, waste
calendar, transit departures) to family members or flatmates without giving
them the owner's password or the ability to change anything (#82). API
tokens (#96) and kiosk displays (#50) need the same kind of limited access.

## Decision

- There is still exactly **one owner** with full rights. The setup, the
  password reset via `trisse-home reset-account` and all settings stay
  owner-only.
- The owner can create up to 20 **guests**: own name and password, own
  sessions and optional two-factor authentication, a list of **shared tabs**,
  and a disabled flag. Guests sign in with name + password; the owner keeps
  signing in with the password only.
- **Authorization is decided centrally** in the auth module's request guard
  (`usecase.Authorizer`, rule `domain.GuestMayAccess`) — controllers of other
  modules contain no permission checks. The rule is an **allow-list**:
  - self-service on the own account (password, two-factor, sessions, logout);
  - `GET` of the dashboard (filtered to shared tabs), modules list and
    appearance;
  - `GET` of modules whose widgets are on a shared tab (resolved through the
    `GuestScope` port, implemented in the composition root), except any
    `/settings` endpoint and the auth, transfer, backup, notification and
    backup-monitor modules;
  - everything else, including every write, is `403`.
  New endpoints are therefore closed for guests by default.
- The authenticated account travels as `kernel.Principal` in the request
  context; modules that show per-user views (the dashboard) read it there.
- Sessions, recovery codes and security events belong to an account
  (`auth_accounts`, migration 0005 moves the owner over unchanged).

## Consequences

- One permission model that API tokens and kiosk tokens can reuse (scopes
  instead of tabs).
- Changing a guest's tabs or disabling it takes effect on the next request:
  the account is loaded per request (one indexed SQLite lookup).
- Widget data is shared at module granularity: a guest who sees the notes
  widget can read all notes, not only those shown in the widget. The
  settings page says so; sensitive data belongs on tabs that are not shared.
- Guests cannot change anything, not even check off a to-do. Shared
  editing would need per-object ownership and is out of scope.
