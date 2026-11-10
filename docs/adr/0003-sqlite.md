# 0003 – SQLite via a pure-Go driver

- Status: Accepted
- Date: 2026-11-10

## Context

A single user stores a modest amount of data (layout, bookmarks, notes, feed
items). Running a separate database server would add operational overhead.

## Decision

Use SQLite through `modernc.org/sqlite` (pure Go, no CGO) in WAL mode with a
single connection. Each module ships embedded SQL migrations
(`adapter/gateway/migrations/NNNN_name.sql`); the platform applies them per
module and records them in `schema_migrations`. Only
`internal/platform/database` imports the driver; gateways use `database/sql`.

## Consequences

- Backups are a single file (`/data/trisse-home.db`, plus WAL while running).
- Write throughput is limited, which is irrelevant for a single user.
- Times are stored as UTC text in a sortable format (`sqlutil.FormatTime`).
