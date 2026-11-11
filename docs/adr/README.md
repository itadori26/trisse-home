# Architecture Decision Records

Each record documents one significant decision: its context, the decision and
its consequences. Records are immutable; a changed decision gets a new record
that supersedes the old one.

| No.                                         | Title                                         | Status   |
| ------------------------------------------- | --------------------------------------------- | -------- |
| [0001](0001-clean-architecture.md)          | Strict Clean Architecture for every module    | Accepted |
| [0002](0002-go-backend.md)                  | Go with the standard library for the backend  | Accepted |
| [0003](0003-sqlite.md)                      | SQLite via a pure-Go driver                   | Accepted |
| [0004](0004-svelte-spa.md)                  | Svelte 5 SPA embedded into the binary         | Accepted |
| [0005](0005-single-user-auth.md)            | Single-user auth with setup token and sessions| Accepted |
| [0006](0006-background-jobs-cache-aside.md) | Background jobs and cache-aside for integrations | Accepted |
| [0007](0007-own-dashboard-grid.md)          | Own grid implementation for the dashboard     | Accepted |
| [0009](0009-two-factor-authentication.md)   | Two-factor authentication with TOTP           | Accepted |
| [0008](0008-sample-data-on-setup.md)        | Sample data after the first setup             | Accepted |
| [0010](0010-internationalisation.md)        | Internationalisation without a framework      | Accepted |
| [0011](0011-lean-by-default.md)             | Lean by default: dependency policy and resource budgets | Accepted |
| [0012](0012-owner-and-guest-roles.md)       | Owner and read-only guest accounts (supersedes the single-user scope of 0005) | Accepted |
