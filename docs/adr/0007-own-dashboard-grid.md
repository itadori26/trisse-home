# 0007 – Own grid implementation for the dashboard

- Status: Accepted
- Date: 2026-11-10

## Context

The dashboard needs a 12-column grid with drag & drop and resizing. At the
time of writing, Svelte grid libraries either did not support Svelte 5 or
were framework-agnostic DOM libraries that fight with Svelte's rendering.

## Decision

Implement the grid ourselves: the layout algorithm (collision detection,
vertical compaction, moving and resizing) is pure TypeScript in
`features/dashboard/domain/layout.ts` and fully unit-tested; the Svelte
component only renders CSS grid positions and translates pointer events.

## Consequences

- No dependency, ~300 lines of code we own and test.
- Fewer features than big grid libraries (no responsive breakpoints per
  layout); on small screens widgets are simply stacked.
