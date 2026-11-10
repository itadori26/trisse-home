# 0004 – Svelte 5 SPA embedded into the binary

- Status: Accepted
- Date: 2026-11-10

## Context

The start page is highly interactive (drag & drop, live widgets) but needs no
server-side rendering. Node should not be required at runtime.

## Decision

The frontend is a Svelte 5 + Vite + TypeScript single-page application. Its
build output is embedded with `go:embed` and served by the Go binary. The
frontend follows the same layering per feature (`domain`, `usecase`,
`adapter/api`, `ui`); `src/app` is its composition root. TypeScript DTO types
are generated from the Go DTOs with tygo and committed.

## Consequences

- Node is only needed at build time (Docker build stage).
- A strict Content-Security-Policy is possible (`script-src 'self'`).
- DTO changes require `make generate`; CI fails on stale generated files.
