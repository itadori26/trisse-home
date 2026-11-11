# Demo mode and screenshots

`DEMO_MODE=true` turns an instance into a public showcase that is safe to
expose to strangers.

## What demo mode does

| Concern | Behaviour |
| --- | --- |
| Login | No password. The browser signs in automatically through `POST /api/auth/demo`; a banner explains the demo. |
| Data | A fresh demo account is created at start and filled with the sample data. A snapshot of all exportable data is restored every `DEMO_RESET_INTERVAL` (default 3 h). |
| Outbound requests | None. Every module's HTTP client is replaced by fixture adapters (`internal/platform/demo`): weather, news feeds and Docker answer with made-up sample data, everything else is refused. Service status checks report "up" without probing. Wake-on-LAN is disabled. |
| Credentials | Never stored: the secret store refuses to encrypt, so integration settings cannot be saved. |
| Account changes | Password, two-factor setup, session revocation, logout, manual backups, data import and the update check are answered with 403. Automatic backups are off. |

The use cases do not know about demo mode: `cmd/server` swaps the adapters
when it wires the modules, exactly like it would for any other gateway.

## Start-up guards

Demo mode refuses to start when

- the database already contains an account that was not created by demo mode
  (i.e. a production database — use an empty data directory), or
- `DOCKER_ACTIONS`, `WOL_RELAY` or `WIREGUARD_DIR` is set (a demo must not
  reach real systems).

## Running a demo

```sh
docker run -d --name trisse-home-demo -p 8090:8090 \
  -e DEMO_MODE=true -e SECRET_KEY="$(openssl rand -base64 32)" \
  -e ALLOWED_CIDRS=0.0.0.0/0 \
  ghcr.io/itadori26/trisse-home:latest
```

Put it behind a reverse proxy with HTTPS (`COOKIE_SECURE=true`,
`TRUSTED_PROXIES=…`, see [reverse-proxy.md](reverse-proxy.md)) and a rate
limit. Do **not** run a public demo on the same host or network as a real
instance, and never port-forward a real instance.

## Hosting decision

There is currently **no official hosted demo**. The project does not run
public infrastructure on the maintainer's home server, on purpose: the start
page is meant to stay inside the LAN. A hosted demo needs a separate, isolated
machine (e.g. a small VPS or a free container platform) that the maintainer
operates; until that exists, the screenshots in the README — regenerated for
every release — serve as the preview, and anyone can start a local demo with
the command above in seconds. When a hosted demo exists, its URL is added to
the README and this section.

## Screenshots

The README images are produced by Playwright from a throw-away demo
instance, so they never contain personal data:

```sh
make build && cd web && npm run screenshots   # writes docs/screenshots/*.png
```

`SCREENSHOT_BASE_URL=http://host:port npm run screenshots` takes them from an
already running demo instead. The release workflow does exactly that with the
freshly published image and attaches the images to the GitHub release.
