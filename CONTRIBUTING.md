# Contributing

Thanks for your interest in trisse-home!

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

## UI texts and translations

Every text shown in the UI comes from the message catalogues in
`web/src/shared/i18n/locales/` (see ADR 0010). When you add or change a text,
add the key to `en.ts` and the German text to `de.ts` and use it with
`t('feature.key', { name })`. `npm run check` fails if a German text is
missing, and a unit test verifies that placeholders match.


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
