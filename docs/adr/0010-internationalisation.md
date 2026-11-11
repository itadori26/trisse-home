# ADR 0010: Internationalisation without a framework

- Status: Accepted
- Date: 2026-11-11

## Context

The UI was English only; German-speaking users asked for German (#25). The
frontend already formats dates and numbers with `Intl`. Adding an i18n
library would add a dependency and a build step for message extraction.

## Decision

- Message catalogues are plain TypeScript objects in
  `web/src/shared/i18n/locales/`: `en.ts` is the source (its keys form the
  `MessageKey` type), `de.ts` is typed `Record<MessageKey, string>`, so a
  missing German text is a compile error (`svelte-check`), and a Vitest
  checks that both catalogues have the same keys and placeholders.
- `t(key, params)` lives in a `.svelte.ts` module and reads the active
  locale from a `$state`, so every template re-renders when the language
  changes. Placeholders use `{name}`.
- Static structures (widget catalogue, settings sections, tour steps) store
  message keys or getters instead of texts.
- The language is part of the appearance settings (`""` = browser
  language, `en`, `de`) and is stored on the server, so it is the same on
  every device. `intlLocale()` makes `toLocale*` follow the UI language.
- Keys are grouped by feature (`weather.*`, `calendar.*`) plus `common.*`
  and `time.*` for shared phrases.

## Consequences

- No runtime dependency; type safety for keys and completeness.
- Messages created by the backend (validation errors) stay English for now;
  translating them needs message codes in the API (follow-up #102).
- New UI text must be added to both catalogues (see CONTRIBUTING.md).
