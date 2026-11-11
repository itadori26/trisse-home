# trisse-home

[![CI](https://github.com/itadori26/trisse-home/actions/workflows/ci.yml/badge.svg)](https://github.com/itadori26/trisse-home/actions/workflows/ci.yml)

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
| **weather**    | Weather and forecast via [Open-Meteo](https://open-meteo.com) (no API key), official DWD warnings and a rain radar widget (Germany) |
| **clock**      | Time and date, optionally in another time zone                                                        |
| **services**   | Tiles for your services with online status, response time history, TLS certificate expiry warning and Wake-on-LAN |
| **bookmarks**  | Links in folders with tags, filterable                                                                |
| **notes**      | Markdown notes, pinnable to the dashboard (sanitised rendering)                                       |
| **todos**      | To-do list with due dates and overdue highlighting                                                    |
| **rss**        | Feed reader (RSS, Atom, JSON Feed) with read state                                                    |
| **calendar**   | Agenda or month view of CalDAV calendars, e.g. Nextcloud (recurring events supported); quick-add of events |
| **system**     | CPU, memory, swap, disks, load, uptime and temperature of the host, with a 24 h history               |
| **docker**     | Container status and health via a read-only Docker socket proxy; optional restart and log view       |
| **adguard**    | AdGuard Home statistics                                                                               |
| **shopping**   | Shopping list stored as tasks in a CalDAV task list, usable by the whole household in their task apps |
| **transit**    | Next public transport departures of a stop with real-time delays (via [Transitous](https://transitous.org)) |
| **waste**      | Waste collection dates from your municipality’s ICS calendar; tomorrow’s bin is highlighted           |
| **wireguard**  | WireGuard peers with name, last handshake and traffic (from a secret-free export on the host)        |
| **fritzbox**   | FRITZ!Box online status, external IP, current throughput and recent calls (missed calls highlighted)  |
| **nextcloud**  | Nextcloud version, your quota and – for admins with the serverinfo app – free space, users and app updates |
| **appearance** | Light/dark/system theme, accent color, background image, custom CSS                                  |
| **backup**     | Automatic, consistent database backups with retention; download in the settings                       |
| **backupmonitor** | Status of your own backup jobs (restic, borg, rsync …) reported via a secret ping URL             |
| **notifications** | Push messages via [ntfy](https://ntfy.sh) for offline services, containers, full disks, expiring certificates and failed backups |
| **onboarding** | Guided tour after the first login, can be restarted from the settings                                 |
| **auth**       | First-run setup, login, session overview with sign-out per device, password change (always enabled)   |

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
git clone https://github.com/itadori26/trisse-home.git
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

**Backup:** the app creates consistent database backups automatically
(daily, 7 kept; see `BACKUP_*`) in `/data/backups` and offers them for
download under Settings → Backups. Keep `SECRET_KEY` separately; without it
stored integration passwords cannot be decrypted. To copy the backups off the
volume:

```sh
sudo docker run --rm -v trisse-home-data:/data -v "$PWD":/backup alpine \
  cp -r /data/backups /backup/
```

**Restore:** stop the stack, replace `/data/trisse-home.db` in the volume
with a backup file (and remove `trisse-home.db-wal`/`-shm` if present), then
start it again.

**Forgot the password?**
`sudo docker compose exec trisse-home /trisse-home reset-account`, then
`sudo docker compose restart trisse-home` and use the new setup token from
the log.

## WireGuard widget (optional)

trisse-home has no access to WireGuard itself. A small script on the host exports the peer status every minute into
`./wireguard`, which is mounted read-only into the container:

```sh
sudo tee /etc/cron.d/trisse-home-wireguard <<'CRON'
* * * * * root /opt/stacks/trisse-home/scripts/wireguard-export.sh
CRON
```

The script runs `wg show all dump` in the `wireguard` container (`WG_CONTAINER=` for WireGuard on the host), keeps
only the peer lines — the interface line with the **private key is dropped and preshared keys are blanked** — and
reads the peer names of `linuxserver/wireguard` from the public keys in its `peer_<name>` folders. Names can be
changed in Settings → WireGuard. A peer counts as online when its last handshake is less than 3 minutes old.

## Wake-on-LAN (optional)

Service tiles with a MAC address get a ⏻ button while the service is not online. Magic packets are broadcasts, and
broadcasts cannot leave a Docker bridge network — so trisse-home stays in its own network and a tiny relay does
the broadcasting. Enable it with one line in `.env`, then run `docker compose up -d` as usual:

```sh
COMPOSE_PROFILES=wol                 # or wol,docker-actions
# WOL_BROADCAST=192.168.178.255:9    # optional, if 255.255.255.255 leaves through the wrong interface
```

The `wol-relay` service uses the same image (`trisse-home wol-relay`), runs in the host network, read-only and
without capabilities, has no access to data, Docker or secrets, accepts nothing but valid magic packets on UDP
port 9009 and is rate-limited. The port is reachable from your LAN, which only allows what every LAN device can
do anyway: wake devices.

| Setup | What to do |
|---|---|
| Docker Compose (default) | `COMPOSE_PROFILES=wol` |
| Binary / container in host network | `WOL_RELAY=` (empty) — trisse-home broadcasts itself |
| No extra container wanted | Enable directed broadcast forwarding on the host (`sysctl net.ipv4.conf.all.bc_forwarding=1` and `net.ipv4.conf.<lan-if>.bc_forwarding=1`), set `WOL_RELAY=` and `WOL_BROADCAST=<lan>.255:9` |

The target device must have Wake-on-LAN enabled in its BIOS/UEFI and network driver.

## Docker actions (optional)

The Docker widget can restart containers and show their last 200 log lines. This is **off by default** because a
restart needs write access to Docker. To enable it, add to `.env`:

```sh
COMPOSE_PROFILES=docker-actions        # starts the second, restart-only socket proxy
DOCKER_ACTIONS=true
DOCKER_ACTIONS_CONTAINERS=homarr,dashdot   # exact container names
```

The read-only proxy stays unchanged. The additional `docker-actions-proxy` only accepts
`POST /containers/<name>/stop|restart|kill` and denies everything else (no reads, no create, no exec), and
trisse-home only calls *restart* for containers on the allow-list. Logs are read through the read-only proxy and
may contain secrets — only list containers whose logs you are fine seeing on the dashboard.

## Notifications and backup monitoring

**Notifications** (Settings → Notifications) publish to an [ntfy](https://ntfy.sh) topic. A state change is only
reported after two consecutive checks (no flapping) and the first observation after a restart only sets the
baseline. Topics on public servers are readable by anyone who knows the name — use a long random topic (one is
suggested) or your own ntfy server with an access token; the token is stored encrypted with `SECRET_KEY`.

**Backup monitoring** (Settings → Backup monitoring) gives every backup job a secret ping URL, shown once. Call it
at the end of your backup script:

```sh
restic backup /data && curl -fsS -X POST "$PING_URL" || curl -fsS -X POST --data "restic failed" "$PING_URL?status=fail"
```

A source is *overdue* when no successful run arrives within its maximum age. The ping endpoint needs no login but
is still subject to `ALLOWED_CIDRS`; only a SHA-256 hash of the token is stored, and a new URL can be generated at
any time.

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
| `TRANSIT_API_URL`   | `https://api.transitous.org`      | MOTIS API for departures; point it to your own MOTIS instance if you like.    |
| `WIREGUARD_DIR`     | `/wireguard` (compose)            | Directory with the export of `scripts/wireguard-export.sh`.                   |
| `WOL_RELAY`         | `host.docker.internal:9009` (compose) | Wake-on-LAN relay; empty sends the broadcast directly (no Docker / host network). |
| `WOL_BROADCAST`     | `255.255.255.255:9`               | Broadcast address the magic packet is sent to (by the relay or directly).     |
| `DOCKER_ACTIONS`    | `false`                           | Enables restart and log view in the Docker widget (see below).                |
| `DOCKER_ACTIONS_CONTAINERS` | empty                     | Containers (exact names) that may be restarted and have their logs shown.     |
| `DOCKER_ACTIONS_PROXY_URL` | `http://docker-actions-proxy:2375` | Restart-only socket proxy.                                         |
| `TZ`                | `UTC`                             | Time zone for logs and floating calendar times.                               |
| `LOG_LEVEL`         | `info`                            | `debug`, `info`, `warn` or `error`.                                           |
| `LOG_FORMAT`        | `json`                            | `json` or `text`.                                                             |
| `DATABASE_PATH`     | `/data/trisse-home.db`            | SQLite file.                                                                  |
| `BACKUP_DIR`        | `backups` next to the database    | Directory for automatic backups.                                              |
| `BACKUP_INTERVAL`   | `24h`                             | How often a backup is created (≥ 1h).                                         |
| `BACKUP_RETENTION`  | `7`                               | Number of backups kept.                                                       |
| `SHUTDOWN_TIMEOUT`  | `15s`                             | Grace period for in-flight requests.                                          |
| `SAMPLE_DATA`       | `true`                            | Fill a fresh installation with example widgets and data after setup.         |
| `SAMPLE_WEATHER_LOCATION` | `Berlin,52.52,13.41`        | `Name,lat,lon` of the example weather widget.                                 |
| `SAMPLE_FEEDS`      | Go Blog, Hacker News              | Example feeds, comma separated.                                              |

### Security model

- IP allowlist before everything else; denied and first-seen allowed client
  IPs are logged so the list can be tuned. Do **not** forward the port on
  your router — use a VPN (e.g. WireGuard) for remote access.
- Single user, Argon2id password hash, login rate limiting (5 attempts per
  15 minutes per IP), optional two-factor authentication (TOTP, RFC 6238)
  with ten one-time recovery codes; the TOTP secret is encrypted at rest,
  recovery codes are stored hashed and reused codes are rejected.
- Session overview under Settings → Account to sign out other devices.
- Session cookie `HttpOnly` + `SameSite=Strict`; CSRF protection via a
  per-session token **and** `Sec-Fetch-Site`/`Origin` checks.
- Strict Content-Security-Policy (`script-src 'self'`), Markdown sanitised
  with DOMPurify, feed links restricted to http(s).
- Integration passwords encrypted at rest and never returned by the API.
- The container runs as non-root on a read-only root file system without
  capabilities. Docker access only through `tecnativa/docker-socket-proxy`
  with container listing enabled and everything else (including all POST
  requests) denied, on an internal network without published port.

### Outbound requests

Several modules fetch URLs that you configure yourself: service health checks,
RSS feeds, CalDAV calendars, AdGuard Home, Nextcloud, the FRITZ!Box and your
ntfy server. This is intended behaviour, so static analysis reports these as
"server-side request forgery" — only the single, authenticated user can
configure them (session + CSRF protection, IP allowlist). Weather data comes
from Open-Meteo, weather warnings and the rain radar from the open data
services of the Deutscher Wetterdienst (only for locations in Germany; the
radar image is proxied, so the browser never contacts DWD) and departures
from the community-run [Transitous](https://transitous.org) API (cached one
minute per stop; please respect its fair-use policy or run your own MOTIS).
Nothing else leaves the server.

The FRITZ!Box login uses HTTP digest authentication, which TR-064 defines with
MD5. Static analysis flags this as weak password hashing; the password is not
stored as a hash (it is encrypted with `SECRET_KEY`), MD5 only answers the
box's login challenge as the protocol requires. The search endpoint redirects to the search engine (or to a URL typed
into the search bar); it requires a session, and the `SameSite=Strict` cookie
is never sent with cross-site requests, so it cannot be abused as an open
redirect.

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
