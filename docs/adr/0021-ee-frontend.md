# ADR 0021: The frontend of `ee/` is a separate plugin package

- Status: Accepted (2026-11-16)
- Date: 2026-11-16
- Issue: [#307](https://github.com/itadori26/waylight/issues/307), part of [#292](https://github.com/itadori26/waylight/issues/292)
- Builds on: [ADR 0017](0017-open-core.md), [ADR 0020](0020-ee-stays-proprietary.md)

## Context

ADR 0017 keeps the Go code of `ee/` out of the core: the core never imports
it, it is wired only in the composition root behind the build tag `ee`, and a
build without the tag is MIT only. It says nothing about the frontend. The
`ee/` features (licence keys, announcements, absences, branding, team tabs)
all have screens, so the web app needs the same separation. People should
also see in the settings which features for organisations exist, including
the ones that are switched off.

## Decision

### A plugin package with its own build

- The frontend of `ee/` is its own package in `ee/web/` with its own
  `package.json` and Vite build. It produces one ES module,
  `plugin.js`, plus its lazy chunks.
- The core web app (`web/`) never imports from `ee/web/`. The package may
  import types and plain helpers from the core's plugin API
  (`web/src/app/plugins/`), nothing else.
- The plugin bundles its own Svelte runtime. Its screens do not become part
  of the core's component tree: the core hands each one a host element and
  the plugin mounts into it and unmounts on request. This keeps the two
  builds independent of each other's compiler and runtime versions.

### The plugin API (core, MIT)

- A plugin exports `plugin: WaylightPlugin` with widgets, settings sections and
  search contributions. Each entry carries its metadata (kind, title,
  category, icon, default size, the licence feature it needs) and a
  `mount(host, context)` function that returns an `unmount` function.
- The context gives the plugin what core features get from the container:
  the API client (session and CSRF handled), the active language, the signed
  in account and its role, and a way to ask the core to reload.
- The core wraps plugin entries into its normal widget, settings and search
  registration, so the picker, the settings dialog and the search field show
  them without knowing about `ee/`.
- Search contributions are commands (`searchActions`, #453): keywords, a
  title and a `run(context)`. The core matches the typed text against the
  keywords and lists the action under *Actions*. Guests get none, and an
  action whose feature is not fully licensed is left out, since it could
  only fail.
- First-start steps (`onboardingSteps`, #671) add questions to the
  first-start flow after the core's own: a title and a view that registers
  what *Next* saves. The core draws the progress and the buttons. Guests get
  none, owner-only steps reach only the owner, and a step whose feature is
  not fully licensed is left out, so it never stores a setting that is not
  shown.
- The contract grows by optional fields only. `apiVersion` goes up when
  fields are added (version 2: compact views of plugin widgets and German
  texts of the plugin's backend messages, #632; version 3: first-start
  steps, #671); the loader accepts every
  version from 1 up to its own, so a plugin built against an older contract
  keeps loading, and an older core ignores fields it does not know.
- A plugin widget can name the source of its compact view (`compact`), and a
  plugin can bring translations of its backend messages (`serverMessages`).
  `make generate` writes the English catalogue of the `ee/` messages into
  `ee/web/src/generated/`, next to the German texts in `ee/web/src/serverDe.ts`,
  so the core's catalogue never lists them.

### Loading

- `GET /api/plugins` (core) lists the plugin bundles the server offers. A
  build without the tag `ee` offers none, and the app loads nothing.
- A build with the tag `ee` embeds `ee/web/dist` and serves it under
  `/ee/`. `GET /api/plugins` then lists `/ee/plugin.js`, and the app loads it
  with a dynamic `import()` after sign-in.
- Releases and images stay built without the tag until the licence files are
  final (#319).

### Features that are switched off stay visible

- The plugin always registers all of its features. Widgets of a feature
  without a licence are left out of the widget picker, as the server leaves
  the module out of `GET /api/modules` and answers its endpoints with 403.
- Settings sections are listed even then. A section whose feature is off
  shows what the feature does and that it needs a licence key, instead of
  its form. One section, *Features for organisations*, lists every feature
  with its state (on, read only, off) from `GET /api/licence/features` and
  takes the licence key.

### Tests and checks

- The plugin package has its own Vitest run in CI, next to the core's.
- An architecture test fails if a file under `web/` imports from `ee/`.
- `make build-ee` builds the plugin and the server with `-tags ee`; CI runs
  it, and an end-to-end test loads the plugin with a test licence key.

## Consequences

- The MIT bundle does not grow with `ee/` features; the plugin is downloaded
  only from servers that offer it.
- A second Svelte runtime is loaded with the plugin, about 20 KB gzipped.
- Plugin screens cannot use core components directly; shared look comes from
  the core's CSS variables, which apply inside the host element.
- The plugin API becomes an interface the core keeps stable for `ee/`.

## Alternatives considered

- **`ee/` frontend inside `web/src/features/`:** simplest, but it puts code
  under another licence into the MIT tree and into every bundle.
- **`ee/web/` compiled into the core bundle behind a build flag:** one
  build with two variants, but the core build would depend on `ee/` sources.
- **Sharing the core's Svelte runtime through an import map:** smaller, but
  ties the plugin to the exact runtime build of the core.
