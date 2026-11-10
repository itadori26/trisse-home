# trisse-home

A self-hosted personal start page in the spirit of iGoogle: freely arranged
widgets on multiple tabs plus small modules for everyday life and the home
server — as lean as possible, as much as necessary, built on a strict Clean
Architecture.

- **One binary, one container (~25 MB):** Go backend with the Svelte frontend
  embedded; SQLite as database. Node is only needed to build.
- **Private by default:** single-user login, IP allowlist, no telemetry, no
  external requests except to the services you configure.
- **Modular:** every feature is a module that can be switched off.

## Features

| Module         | What it does                                                                                          |
| -------------- | ----------------------------------------------------------------------------------------------------- |
| **dashboard**  | Tabs (e.g. Home / Server / Study) with a 12-column drag & drop grid; layout stored in the database    |
| **search**     | Search bar with DuckDuckGo as default (all `!bangs` work), other engines, custom shortcuts, OpenSearch |
| **weather**    | Current weather and forecast via [Open-Meteo](https://open-meteo.com) (no API key), place search      |
| **clock**      | Time and date, optionally in another time zone                                                        |
| **services**   | Tiles for your services with online status (periodic health checks)                                  |
| **bookmarks**  | Links in folders with tags, filterable                                                                |
| **notes**      | Markdown notes, pinnable to the dashboard (sanitised rendering)                                       |
| **todos**      | To-do list with due dates and overdue highlighting                                                    |
| **rss**        | Feed reader (RSS, Atom, JSON Feed) with read state                                                    |
| **calendar**   | Upcoming events from CalDAV calendars, e.g. Nextcloud (recurring events supported)                    |
| **system**     | CPU, memory, swap, disks, load, uptime and temperature of the host                                    |
| **docker**     | Container status and health via a read-only Docker socket proxy                                       |
| **adguard**    | AdGuard Home statistics                                                                               |
| **appearance** | Light/dark/system theme, accent color, background image, custom CSS                                  |
| **auth**       | First-run setup, login, sessions, password change (always enabled)                                    |

## Architecture in brief

Every backend module is split into the four rings of Clean Architecture;
source dependencies only point inwards. The rule is enforced by `depguard`
(golangci-lint) and an architecture test, and the frontend follows the same
layering per feature. Details: [docs/architecture.md](docs/architecture.md)
and the [ADRs](docs/adr/README.md).

```
cmd/server/              composition root (wiring only)
internal/
  platform/              frameworks & drivers: config, db, http server, middleware, scheduler, crypto …
  shared/                kernel (error kinds, Clock, IDGenerator), HTTP and SQL adapter helpers
  modules/<name>/
    domain/              entities, value objects, rules           (ring 1)
    usecase/             use cases, input/output ports, models    (ring 2)
    adapter/controller/  HTTP handlers                            (ring 3)
    adapter/presenter/   response DTOs
    adapter/gateway/     SQLite repositories, API clients, migrations
    module.go            wires the module, implements modules.Module
web/src/
  app/                   frontend composition root, widget registry
  features/<name>/{domain,usecase,adapter/api,ui}/
  shared/                API client, generated DTO types, theme, UI
api/openapi.yaml         API description
```

## Quick start (Docker Compose)

Requirements: Docker with the Compose plugin.

```sh
git clone https://github.com/<you>/trisse-home.git
cd trisse-home
cp .env.example .env
sed -i "s|^SECRET_KEY=.*|SECRET_KEY=$(openssl rand -base64 32)|" .env
# optional: set BIND_ADDRESS to your server's LAN IP and adjust ALLOWED_CIDRS
sudo docker compose up -d --build
sudo docker compose logs trisse-home | grep setup_token
```

Open `http://<server-ip>:8090`, enter the setup token from the log and choose
a password (at least 12 characters). The token is only valid until the
account exists.

After the setup the start page is filled with **sample data**: a *Home* tab
with search, clock, weather, to-dos, a welcome note, bookmarks, news and
calendar, and a *Server* tab with service tiles, host metrics, Docker and
AdGuard Home. Delete or change whatever you like; set `SAMPLE_DATA=false` to
start with an almost empty page instead.

**Updating:** `git pull && sudo docker compose build && sudo docker compose up -d`

**Backup:** the data lives in the Docker volume `trisse-home-data`
(`/data/trisse-home.db`). Keep `SECRET_KEY` together with the backup; without
it stored integration passwords cannot be decrypted.

```sh
sudo docker run --rm -v trisse-home-data:/data -v "$PWD":/backup alpine \
  tar czf /backup/trisse-home-$(date +%F).tgz -C /data .
```

**Forgot the password?**
`sudo docker compose exec trisse-home /trisse-home reset-account`, then
`sudo docker compose restart trisse-home` and use the new setup token from
the log.

## Configuration

All settings are environment variables (see the fully commented
[.env.example](.env.example)). Integrations such as AdGuard Home, CalDAV and
service tiles are configured in the web interface.

| Variable            | Default                           | Description                                                                  |
| ------------------- | --------------------------------- | ---------------------------------------------------------------------------- |
| `SECRET_KEY`        | — (required)                      | 32 random bytes, base64. Encrypts stored credentials (AES-256-GCM).           |
| `BIND_ADDRESS`      | `0.0.0.0`                         | Host address the port is published on (compose only).                        |
| `HOST_PORT`         | `8090`                            | Host port (compose only).                                                     |
| `LISTEN_ADDR`       | `:8090`                           | Listen address inside the container.                                          |
| `ALLOWED_CIDRS`     | loopback + private ranges         | Client networks allowed to connect. Loopback is always allowed.               |
| `TRUSTED_PROXIES`   | empty                             | Reverse proxies whose `X-Forwarded-For` header is trusted.                    |
| `COOKIE_SECURE`     | `false`                           | `true` once served via HTTPS: adds `Secure` and the `__Host-` cookie prefix.  |
| `SESSION_TTL`       | `720h`                            | Sliding session lifetime.                                                     |
| `SETUP_TOKEN`       | random, logged                    | Fixed setup token (≥ 16 characters) for automated setups.                     |
| `MODULES_DISABLED`  | empty                             | Comma-separated modules to switch off, e.g. `docker,adguard`.                 |
| `SYSTEM_DISK_PATHS` | `System=/host/rootfs` (compose)   | Disks for the system widget as `Label=path`, comma separated.                 |
| `HOST_PROC`         | `/proc` (`/host/proc` in compose) | Host procfs for the system module.                                            |
| `HOST_SYS`          | `/sys` (`/host/sys` in compose)   | Host sysfs (temperatures).                                                    |
| `DOCKER_PROXY_URL`  | `http://docker-socket-proxy:2375` | Docker API endpoint (the bundled read-only proxy).                            |
| `TZ`                | `UTC`                             | Time zone for logs and floating calendar times.                               |
| `LOG_LEVEL`         | `info`                            | `debug`, `info`, `warn` or `error`.                                           |
| `LOG_FORMAT`        | `json`                            | `json` or `text`.                                                             |
| `DATABASE_PATH`     | `/data/trisse-home.db`            | SQLite file.                                                                  |
| `SHUTDOWN_TIMEOUT`  | `15s`                             | Grace period for in-flight requests.                                          |
| `SAMPLE_DATA`       | `true`                            | Fill a fresh installation with example widgets and data after setup.         |
| `SAMPLE_WEATHER_LOCATION` | `Berlin,52.52,13.41`        | `Name,lat,lon` of the example weather widget.                                 |
| `SAMPLE_FEEDS`      | Go Blog, Hacker News              | Example feeds, comma separated.                                              |

### Security model

- IP allowlist before everything else; denied and first-seen allowed client
  IPs are logged so the list can be tuned. Do **not** forward the port on
  your router — use a VPN (e.g. WireGuard) for remote access.
- Single user, Argon2id password hash, login rate limiting (5 attempts per
  15 minutes per IP).
- Session cookie `HttpOnly` + `SameSite=Strict`; CSRF protection via a
  per-session token **and** `Sec-Fetch-Site`/`Origin` checks.
- Strict Content-Security-Policy (`script-src 'self'`), Markdown sanitised
  with DOMPurify, feed links restricted to http(s).
- Integration passwords encrypted at rest and never returned by the API.
- The container runs as non-root on a read-only root file system without
  capabilities. Docker access only through `tecnativa/docker-socket-proxy`
  with container listing enabled and everything else (including all POST
  requests) denied, on an internal network without published port.

### Behind a reverse proxy (optional)

The app is prepared for HTTPS behind a reverse proxy such as
[nginx-proxy](https://github.com/nginx-proxy/nginx-proxy) with acme-companion;
see [docs/reverse-proxy.md](docs/reverse-proxy.md) and
[compose.proxy.example.yaml](compose.proxy.example.yaml). In short: join the
proxy's network, set `VIRTUAL_HOST`/`LETSENCRYPT_HOST`, set
`TRUSTED_PROXIES` to the proxy network and `COOKIE_SECURE=true`.

## Development

Requirements: Go 1.27+, Node 22+ (26 LTS recommended), make,
[golangci-lint](https://golangci-lint.run) v2 and
[tygo](https://github.com/gzuidhof/tygo) (`go install github.com/gzuidhof/tygo@latest`).

```sh
make deps        # Go modules and npm packages
make generate    # TypeScript DTO types from the Go DTOs
make lint        # golangci-lint (incl. depguard), ESLint, Prettier, svelte-check
make test        # Go tests (race detector) and Vitest
make build       # frontend + single binary in bin/
make run         # run locally with ./data (uses .env)
make smoke       # Playwright smoke test (once: cd web && npx playwright install chromium)
make cover       # Go coverage summary
make help        # all targets
```

Frontend development with hot reload: run the backend (`make run`), then
`cd web && npm run dev` — Vite proxies `/api` to `http://127.0.0.1:8090`.

### Tests

| Layer          | How                                                                     |
| -------------- | ----------------------------------------------------------------------- |
| Domain         | Table-driven unit tests, no dependencies                                |
| Use cases      | Fakes for all output ports — no database, no HTTP                       |
| Gateways       | Integration tests against a temporary SQLite / `httptest` servers       |
| Controllers    | `httptest` with fake use cases                                          |
| API end to end | `cmd/server/api_test.go`: all modules wired, fake upstream services     |
| Architecture   | `internal/archtest` and depguard                                        |
| Frontend       | Vitest for domain and use cases of every feature                        |
| Smoke          | Playwright: setup → dashboard → logout → login → edit mode              |

## How to add a new module

Example: a `quotes` module that shows a random quote.

1. **Domain** – `internal/modules/quotes/domain/`: `doc.go` (stating the
   ring), the entity `Quote` with a constructor that enforces its
   invariants, and domain errors wrapping kinds from
   `internal/shared/kernel` (e.g. `kernel.NewPublicError(kernel.ErrNotFound, "…")`).
   Only the standard library and the kernel may be imported. Add table tests.
2. **Use cases** – `internal/modules/quotes/usecase/`: one input-port
   interface per use case (e.g. `RandomQuote`), the output ports it needs
   (e.g. `QuoteRepository`, `kernel.Clock`), request/response models and the
   implementations. Never return entities — map them to response models.
   Test with fakes.
3. **Gateway** – `adapter/gateway/`: implement the output ports (SQLite via
   `database/sql`, HTTP clients injected from outside). Put migrations in
   `adapter/gateway/migrations/0001_init.sql`, embed them and expose
   `Migrations() fs.FS`. Map persistence rows ↔ entities and write
   integration tests with `dbtest.Open`.
4. **Presenter** – `adapter/presenter/`: JSON DTOs with `json` tags and
   mapping functions from the response models.
5. **Controller** – `adapter/controller/`: request DTOs in `dto.go` and
   handlers that decode with `httpapi.DecodeJSON`, call the input port and
   answer with `httpapi.WriteJSON` / `httpapi.WriteDomainError`. Test with
   `httptest` and fake use cases.
6. **Module** – `internal/modules/quotes/module.go`: embed `modules.Base`,
   build gateway → use cases → handlers in `New(...)` and implement `Name`,
   `Migrations`, `RegisterRoutes` (use `r.Protected` for everything that
   needs a session), `Jobs` and `WidgetKinds` as needed. Optionally
   implement `modules.Seeder` with an idempotent sample-data use case and
   add the widget to `SampleLayout` in the dashboard module.
7. **Composition root** – add the module to the `optional` list in
   `cmd/server/modules.go` and to the module list in `.env.example`.
8. **Types** – add the presenter (and controller `dto.go`) packages to
   `tygo.yaml` and run `make generate`.
9. **Frontend** – `web/src/features/quotes/` with `domain/`, `usecase/`
   (port interface and use case class), `adapter/api/` (implements the port,
   maps DTO → domain), `ui/` (widget and optional settings component) and an
   `index.ts` exporting the public API. Wire it in
   `web/src/app/container.ts` and register the widget in
   `web/src/app/widgets.ts`.
10. **Docs** – add the endpoints to `api/openapi.yaml`, mention the module in
    this README and write an ADR if the module introduces a notable
    decision. Run `make ci` before committing.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: see
[SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE)
