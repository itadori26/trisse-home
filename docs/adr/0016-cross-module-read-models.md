# ADR 0016: Cross-module read models

- Status: Accepted
- Date: 2026-11-12
- Context issues: #52 (daily briefing), #70 (briefing push)

## Context

Until now every module was self-contained: the dependency rule
([ADR 0001](0001-clean-architecture.md)) forbids a module from importing
another one, enforced by `depguard` and `internal/archtest`. The only
cross-module traffic were shared kernel ports that many modules *write* to
(observations, alerts, audit events).

The daily briefing (#52) is the first feature that *reads* from several
modules at once: due to-dos, upcoming appointments, weather warnings, waste
collections, services that are down and overnight incidents from the
timeline. It must keep working when any of these modules is disabled via
`MODULES_DISABLED`, and the source modules must not learn that a briefing
exists.

Options considered:

1. **Import the source modules' use case packages from the briefing.**
   Simple, but couples the briefing to six modules' internals and breaks the
   rule that modules do not import each other.
2. **Add shared read ports to `internal/shared/kernel`.** Every source module
   would have to implement a kernel interface designed for one consumer; the
   kernel would grow with every new read model.
3. **Consumer-owned ports, adapters in the composition root.** The briefing
   defines exactly the narrow read ports it needs, with its own small fact
   types. `cmd/server` implements them on top of public read methods of the
   source modules.

## Decision

Option 3.

- The **consumer owns the ports.** `briefing/usecase` defines one read port
  per source (`DueTodosReader`, `UpcomingEventsReader`,
  `WeatherWarningReader`, `CollectionReader`, `ServiceStatusReader`,
  `IncidentReader`) with briefing-owned fact types (`TodoFact`, …). The
  module knows nothing about to-do IDs, CalDAV or DWD.
- **Source modules expose read methods on their `Module`** (`OpenTodos`,
  `Upcoming`, `Warnings`, `Services`, `Recent`) that delegate to their
  existing input ports. They contain no logic and no knowledge of the
  consumer.
- **Adapters live in the composition root** (`cmd/server/briefing.go`).
  They translate between the source's view types and the consumer's facts
  and do simple filtering (e.g. "offline" services, non-info timeline
  events). The briefing module is built after the source modules and finds
  them in the registry; a disabled module leaves its port `nil`, and the
  use case leaves that section out.
- **Failures stay local:** a source that errors marks its section as
  unavailable; the briefing itself still succeeds.
- **Delivery channels are adapters too:** the daily push (#70) reuses the
  same `BuildBriefing` composition and hands the text to a `Publisher`
  port, implemented in the composition root on top of the notifications
  module.

## Consequences

- No module imports another module; `archtest` stays unchanged.
- Adding a section means: a port and a fact type in the consumer, a read
  method on the source module (if none exists), and an adapter in
  `cmd/server` with an end-to-end test (`cmd/server/briefing_test.go`).
- The composition root grows a little logic (mapping and filtering). It
  must stay mapping only; rules belong in the consumer's domain
  (`briefing/domain`: ordering, priorities, time windows).
- Future read models (e.g. a kiosk overview or config as code, #88) follow
  the same pattern.
