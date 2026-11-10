# 0008 – Sample data after the first setup

- Status: Accepted
- Date: 2026-11-10

## Context

An empty start page does not show what the application can do; new users
would have to discover every widget and setting on their own.

## Decision

When the account is created, the auth module notifies a `SetupObserver`
output port. The composition root implements it by calling every module
that implements the optional `modules.Seeder` interface. Each module seeds
only its own data through an idempotent use case (it does nothing if the
module already has data), so modules stay independent and a reset followed
by a second setup creates no duplicates. Example feeds and service tiles
are fetched/checked in the background so the setup request stays fast.
`SAMPLE_DATA=false` disables seeding; `SAMPLE_WEATHER_LOCATION` and
`SAMPLE_FEEDS` adapt the examples.

## Consequences

- The first login shows two populated tabs (Home, Server).
- Sample feeds and service tiles cause outbound requests to public sites
  right after setup; operators who do not want that disable sample data.
- New modules can contribute sample data without touching other modules.
