# 0001 – Strict Clean Architecture for every module

- Status: Accepted
- Date: 2026-11-10

## Context

trisse-home bundles many small, independent features (bookmarks, weather,
RSS, server monitoring …). Several of them talk to external systems whose
APIs change over time. The project is also meant as a readable reference that
others can extend.

## Decision

Every backend module is split into the four rings of Clean Architecture —
`domain` (entities), `usecase` (application logic with input and output
ports), `adapter` (controllers, presenters, gateways) and the platform
(frameworks & drivers) — even when a module is small. Source dependencies
only point inwards. Mapping happens explicitly at every boundary
(persistence model ↔ entity ↔ use case model ↔ DTO). Modules cooperate only
through public ports and are wired in the composition root.

The rule is enforced automatically:

- `depguard` (golangci-lint) forbids imports per ring (e.g. `net/http` in the
  domain, the platform in use cases).
- `internal/archtest` parses all imports and additionally forbids imports
  between modules and from adapters into the platform.
- ESLint `no-restricted-imports` applies the same layering to the frontend.

## Consequences

- More files and explicit mapping code per feature; simple CRUD modules look
  verbose. We accept this for uniformity and testability.
- Use cases are tested with fakes, without database or HTTP.
- External services can be replaced by writing a new gateway only.
- Gateway integration tests may import the SQLite driver; `_test.go` files
  are therefore excluded from the adapter→platform depguard rule.
