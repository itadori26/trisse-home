# ADR 0011: Lean by default (dependency policy and resource budgets)

- Status: Accepted
- Date: 2026-11-11

## Context

A small footprint is one of the reasons to choose trisse-home: one static
binary, SQLite, an image of a few megabytes, low idle memory on a home
server that also runs other things. With more than eighty planned features,
every widget adds code, every integration a dependency and every view
requests. Without an explicit rule the footprint grows a little with every
change and nobody notices.

ADR 0001 (Clean Architecture) and `internal/archtest` guard the *direction*
of dependencies. Nothing guarded their *weight*.

## Decision

1. **Dependency policy.** A new runtime dependency (Go module or npm
   `dependencies`) needs a written justification in the pull request: what it
   costs (size, transitive dependencies, maintenance and security track
   record) and why the standard library or a few lines of own code are not
   enough. Development-only tooling (`devDependencies`, linters, generators)
   is exempt.
2. **Defaults for new code.**
   - Standard library first; no icon fonts, no CDNs, no web fonts.
   - Lazy-load everything that is not needed for the first paint (editors,
     QR codes, the inactive locale, rarely used widgets).
   - Aggregate instead of storing raw data long-term (e.g. hourly averages
     for metrics, bounded history tables).
   - Poll only what is visible, cache in the use case (ADR 0006), share
     timers instead of one per widget.
3. **Resource budgets.** CI measures the footprint on every pull request and
   fails when a budget in [`budgets.json`](../../budgets.json) is exceeded
   (`scripts/budgets`, job *docker*). The job summary shows each value, its
   budget and the delta to `main`.

   | Metric | Baseline (2026-11-11) | Budget |
   | --- | ---: | ---: |
   | Initial JS (gzip) | 134 KB | 140 KB (goal ≤ 90 KB after lazy loading, #116) |
   | Initial CSS (gzip) | 6.6 KB | 10 KB |
   | Largest lazy JS chunk (gzip) | – | 60 KB |
   | Go binary | 16.2 MB | 18.9 MB (raised from 17.9 MB for new modules, #174) |
   | Container image | 7.7 MB | 8.5 MB |
   | Idle memory (demo instance, RSS after 60 s) | 76.2 MB | 84 MB |

   Budgets start at baseline + 10 %. **Changing a budget** is a normal pull
   request that edits `budgets.json` and explains why the increase is worth
   it; lowering a budget after an optimisation is encouraged.

## Consequences

- Contributors see the cost of a change in the PR itself instead of months
  later; reviewers have a number to discuss instead of a feeling.
- Some features take more effort (lazy loading, own small helpers instead of
  a library). This is the intended trade-off.
- The budgets are measured on the CI runner; memory varies slightly between
  runs, which the 10 % margin absorbs.
- The PR template asks for "no new dependency, or justified; budgets green",
  and the feature form asks for the expected size and performance impact.
