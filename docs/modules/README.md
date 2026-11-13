# Module documentation

One file per module, `docs/modules/<name>.md` (#193): what it does, its
configuration, the external requests it makes and its API. `make new-module`
creates the skeleton. New modules are described here instead of in the shared
tables of the [README](../../README.md), so two modules in development never
edit the same lines.

## Why modules do not touch shared files

A module consists of files of its own; collectors find them by convention:

| Concern | The module's file | Collected by |
|---|---|---|
| Backend registration | `cmd/server/module_<name>.go` (`registerModule` in `init()`) | `cmd/server/registry.go` |
| TypeScript DTOs | `internal/modules/<name>/adapter/{presenter,controller}` | `scripts/tygoconfig` |
| Backend messages | generated `generated/server-messages/<name>.ts`, German in `i18n/locales/server-de/<name>.ts` | `scripts/servermessages`, `server-de/index.ts` |
| UI texts | `web/src/features/<name>/i18n/{en,de}.ts` | `web/src/app/messages.{en,de}.ts` |
| Widget | `web/src/app/widgets/<name>.ts` | `web/src/app/widgets.ts` |
| Container service | `web/src/app/services/<name>.ts` | `web/src/app/container.ts` |
| Settings section | `web/src/app/settings/<name>.ts` | `web/src/app/SettingsDialog.svelte` |
| OpenAPI | `api/modules/<name>.yaml` (tags, paths, schemas) | `scripts/openapi` (`make openapi`) |
| End-to-end check | `cmd/server/api_<name>_test.go` (`registerModuleE2E`: fake upstream, config override, check) | `cmd/server/e2e_registry_test.go` |
| Documentation | `docs/modules/<name>.md` (what it does, configuration, outbound requests, API) | the directory listing; the README no longer lists modules |
| Search answers (optional) | `web/src/app/search/<name>.ts` (`defineSearch`: intents as data plus an answer source) | `web/src/app/searchSources.ts` |

Existing modules that were added before this layout (the older entries in
`modules.go`, `widgets.ts`, `container.ts` and the shared catalogues) move over
whenever their code is next touched; both styles work side by side.

## Files that are still shared

Only a few files remain that a module may have to touch, and only in a narrow case:

- `internal/platform/config/config.go` and `.env.example`, **only** for a new
  environment variable. Insert the line at the position that matches the
  module's name instead of at the end. Per-module config loading was evaluated
  for #194 and left out: most modules need no variable, and a registration
  indirection for the few that do would cost more than the occasional one-line
  conflict.
- `cmd/server/modules.go`, the `widgets.ts`, `container.ts` and catalogue files
  of older modules, until their code is next touched (they move to the
  per-module files then).

`README.md` and `api/openapi.base.yaml` are not edited for a module any more:
the module is described in `docs/modules/<name>.md` and `api/modules/<name>.yaml`
(`make openapi` merges and validates the fragments into `dist/openapi.yaml`, a
build artifact of every CI run). A test (`TestSelfRegisteredModulesAreDocumented`)
fails when a self-registered module lacks either file.

When a generated file conflicts anyway: never merge it by hand, run
`make generate` instead.
