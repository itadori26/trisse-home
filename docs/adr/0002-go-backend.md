# 0002 – Go with the standard library for the backend

- Status: Accepted
- Date: 2026-11-10

## Context

The application runs on a small home server next to many other containers. It
should start fast, use little memory and be easy to deploy.

## Decision

The backend is written in Go and relies on the standard library wherever
possible: `net/http` with the Go 1.22+ pattern router, `log/slog`,
`http.CrossOriginProtection`, `embed`. Third-party dependencies are limited
to what the standard library cannot provide sensibly: the SQLite driver,
Argon2id (`golang.org/x/crypto`) and an RSS/Atom parser.

## Consequences

- One static binary without CGO, a distroless runtime image of ~8 MB (measured, see ADR 0011).
- No web framework to learn or upgrade; middleware are plain decorators.
- Some helpers (JSON envelope, migrations) are written by hand.
