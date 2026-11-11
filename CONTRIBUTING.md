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
