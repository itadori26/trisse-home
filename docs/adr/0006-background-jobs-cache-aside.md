# 0006 – Background jobs and cache-aside for integrations

- Status: Accepted
- Date: 2026-11-10

## Context

Widgets show data from external systems (Open-Meteo, CalDAV, AdGuard Home,
Docker, service health checks, RSS feeds). Fetching them on every page load
would make the dashboard slow and hammer those services.

## Decision

Each integration module registers a `modules.Job`; the platform scheduler
runs it immediately at start and then periodically. Jobs write results to a
cache (in memory, or SQLite for RSS items). HTTP handlers only read the
cache. When a key is missing (e.g. a newly added weather location), the use
case loads it once synchronously and stores it (cache-aside), after which the
job keeps it fresh.

## Consequences

- Page loads are fast and independent of external latency.
- Data may be a few minutes old; each response carries `updatedAt`.
- Errors of external systems are stored with the cache entry and shown in
  the widget instead of failing the request.
