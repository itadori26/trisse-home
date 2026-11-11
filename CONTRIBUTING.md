# Contributing

Thanks for your interest in trisse-home! Please follow our
[Code of Conduct](CODE_OF_CONDUCT.md). Questions and ideas are welcome in
[Discussions](https://github.com/itadori26/trisse-home/discussions); planned
work is in the [roadmap](ROADMAP.md), and issues labelled
[good first issue](https://github.com/itadori26/trisse-home/labels/good%20first%20issue)
are a good place to start.

## Development setup (about 10 minutes)

Requirements: Go (version in `go.mod`), Node.js 22 or newer, `make`, and
Docker only if you want to build the image.

```sh
git clone https://github.com/itadori26/trisse-home.git && cd trisse-home
make deps          # Go modules, npm packages and the code generators
make ci            # lint, architecture tests, unit tests, type check, build
cp .env.example .env && sed -i "s|^SECRET_KEY=.*|SECRET_KEY=$(openssl rand -base64 32)|" .env
make run           # terminal 1: backend on :8090, data in ./data
cd web && npm run dev   # terminal 2: Vite with hot reload, proxies /api to :8090
```

Open the URL printed by Vite, enter the setup token from the backend log and
pick a password; the sample data fills the dashboard. Add `DEMO_MODE=true` to
`.env` to skip the setup entirely (see [docs/demo.md](docs/demo.md)). The
end-to-end smoke test runs with `cd web && npx playwright install chromium`
once and then `make smoke`.

## Rules

1. Read [docs/architecture.md](docs/architecture.md). Every change must
   respect the dependency rule; `make lint` and `make test` enforce it.
2. Every change starts with an issue (use the issue forms). Pull requests
   reference it with `Closes #<number>`. Keep changes small and focused. New features usually belong in a module
   (see "How to add a new module" in the README).
3. Code, comments and documentation are written in English. Exported Go
   identifiers get GoDoc comments starting with their name; comments explain
   *why*, not *what*.
4. Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
   (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:` …).
5. Run `make ci` before opening a pull request. If you change a DTO, run
   `make generate` and commit the generated TypeScript types.
6. Significant decisions get an ADR in `docs/adr/`.

## Adding a module

Scaffold a complete module with one command:

```sh
make new-module NAME=recipes TITLE="Recipes"
```

This creates `internal/modules/recipes` (domain, use case, gateway with a
migration, presenter, controller, module wiring, all with tests) and
`web/src/features/recipes` (domain, use case, API adapter, widget, tests),
registers the module in `cmd/server/modules.go`, the frontend container and
widget list, `tygo.yaml` and both message catalogues, and regenerates the
TypeScript types. The result passes `make ci` unchanged; then replace the
example `Item` entity with your own model. The insertion points are marked
with `scaffold:` comments — keep them when editing those files.

## Adding an integration

Integrations (Home Assistant, Pi-hole, Jellyfin, …) are tracked in
[#91](https://github.com/itadori26/trisse-home/issues/91). Every integration
follows the same pattern:

1. **Scaffold** a module with `make new-module NAME=pihole`.
2. **Domain types independent of the vendor:** model what the widget shows
   (e.g. `DNSStats` with queries and blocked share), not the vendor's JSON.
3. **Provider port** in the use case ring (e.g. `StatsProvider`); the vendor
   client is a gateway adapter implementing it. A second vendor (AdGuard Home
   vs. Pi-hole) is then just another adapter.
4. **Opt-in outbound requests:** nothing is contacted until the user saves a
   URL in the settings. Use the shared HTTP client passed to `New`, set
   timeouts, cache results in the use case and add the host to the README
   section *Outbound requests*.
5. **Encrypted secrets:** API keys and passwords go through the `Cipher`
   (`SECRET_KEY`, AES-256-GCM) and are only ever returned as "set / not set".
6. **Demo:** if the widget should look good in the demo, add a fixture to
   `internal/platform/demo/transport.go`.

Anything that only needs to show a value from a JSON endpoint does not need
its own module: use the generic JSON widget
([#48](https://github.com/itadori26/trisse-home/issues/48)).

## UI texts and translations

Every text shown in the UI comes from the message catalogues in
`web/src/shared/i18n/locales/` (see ADR 0010). When you add or change a text,
add the key to `en.ts` and the German text to `de.ts` and use it with
`t('feature.key', { name })`. `npm run check` fails if a German text is
missing, and a unit test verifies that placeholders match.

**Adding a language** (e.g. French):

1. Copy `web/src/shared/i18n/locales/de.ts` to `fr.ts`, rename the export and
   translate the values (the type `Record<MessageKey, string>` makes the
   compiler report every missing key).
2. Register it in `catalogs` and `intlLocale` in
   `web/src/shared/i18n/i18n.svelte.ts` and add the option to the language
   select in the appearance settings.
3. Allow the code in `internal/modules/appearance/domain` (`Languages`) and
   extend the completeness test in `i18n.test.ts`.


## Releasing

Releases follow [Semantic Versioning](https://semver.org) and are cut by the
maintainer by pushing a tag on `main`:

```sh
git tag -a v1.2.0 -m v1.2.0 && git push origin v1.2.0
```

The release workflow builds `linux/amd64` and `linux/arm64` images, pushes
`ghcr.io/itadori26/trisse-home:1.2.0`, `:1.2` and `:latest`, smoke-tests both
architectures, attaches an SPDX SBOM and publishes release notes generated
from the Conventional Commit messages (`scripts/release-notes.sh`, also usable
locally to preview the notes).
