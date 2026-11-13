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
| Search answers (optional) | `web/src/app/search/<name>.ts` (`defineSearch`: intents as data plus an answer source) | `web/src/app/searchSources.ts` |

Existing modules that were added before this layout (the older entries in
`modules.go`, `widgets.ts`, `container.ts` and the shared catalogues) move over
whenever their code is next touched; both styles work side by side.

## Files that are still shared

These still see edits from many modules. Keep changes small and insert at the
position that matches the module's name instead of at the end, so two branches
rarely touch the same lines:

- `api/openapi.yaml` (tag, paths, schemas) and the tables of the `README.md`
- `internal/platform/config/config.go`, `.env.example` (only for new
  environment variables)
- `cmd/server/api_test.go` (fake upstream and end-to-end checks)

When such a file conflicts anyway: generated files are never merged by hand,
run `make generate` instead.
