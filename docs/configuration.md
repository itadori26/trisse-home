# Configuration reference

Waylight reads its configuration from environment variables when it
starts. Everything that only you as the signed-in owner should decide, such
as service tiles, calendars or notification channels, lives in the web
interface instead. Settings that widen what the server may do, such as
restarting containers, are deliberately environment-only, so a stolen browser
session cannot switch them on.

Where the variables go depends on how you installed Waylight:

| Installation | File | Apply changes with |
| --- | --- | --- |
| Docker Compose | `.env` next to `compose.yaml` | `sudo docker compose up -d` |
| Debian package, Raspberry Pi image | `/etc/waylight/waylight.env` | `sudo systemctl restart waylight` |

The fully commented [`.env.example`](https://github.com/Waylight-App/waylight/blob/main/.env.example)
is the starting point. Only `SECRET_KEY` is required; the package and the Pi
image generate it on the first start.

Waylight checks the whole configuration at start-up and reports every
problem at once instead of failing on the first one.

Durations use Go syntax: `90s`, `15m`, `24h`, `720h`.

## Essentials

| Variable | Default | Description |
| --- | --- | --- |
| `SECRET_KEY` | none, required | 32 random bytes, base64 encoded (`openssl rand -base64 32`). Encrypts stored passwords and tokens of integrations with AES-256-GCM. Keep a copy: without it, stored credentials cannot be decrypted. |
| `TZ` | `UTC` | Time zone for logs, calendar times without a zone and tab rules, as an IANA name such as `Europe/Berlin`. |
| `MODULES_DISABLED` | empty | Comma-separated modules to switch off, for example `docker,adguard`. Module names are listed in the [feature overview](features.md). |
| `CURRENCY_RATES` | `true` | Currency conversion in the search with the ECB's daily reference rates (fetched at most once a day). `false` for installations without internet. See [markets](modules/markets.md#currency-conversion-in-the-search). |
| `SAMPLE_DATA` | `true` | Fill a fresh installation with example tabs, widgets and data after setup. The template choice on first start replaces the example tabs unless the user keeps them. Existing data is never touched. |
| `CONFIG_FILE` | – | Path to a `waylight.yaml` that is applied at every start. The sections it contains become read-only in the UI. See [Settings as a YAML file](config-as-code.md#managed-mode). |

## Network and access

| Variable | Default | Description |
| --- | --- | --- |
| `BIND_ADDRESS` | `0.0.0.0` | Host address the port is published on. Docker Compose only. |
| `HOST_PORT` | `8090` | Host port. Docker Compose only. |
| `LISTEN_ADDR` | `:8090` | Address the server listens on inside the container or on the host. |
| `ALLOWED_CIDRS` | loopback and private ranges | Client networks that may connect. Loopback is always allowed. Rejected and first-seen addresses are logged, which helps to tune the list. |
| `TRUSTED_PROXIES` | empty | Reverse proxies whose `X-Forwarded-For` header is trusted. Needed behind a proxy so that the allow-list, rate limits and network tab rules see the real client address. |
| `PUBLIC_URL` | empty | The public `https://` address, such as `https://home.example.org`. Required for passkeys and for signing in to Microsoft To Do, Google Tasks and TickTick ([to-do lists](modules/todos.md)) and to Google Calendar and Microsoft 365 ([calendar](modules/calendar.md)), and for password reset links by e-mail ([accounts](accounts.md)). Signing in with Home Assistant in the setup assistant uses it as the client ID ([dashboard](dashboard.md#signing-in-instead-of-pasting-a-key)). |
| `PUSH_RELAY_URL` | `https://push.getwaylight.app` | Where push notifications for the mobile app go ([notifications](notifications.md#waylight-app)). `off` turns the app channel off; nothing is sent to the relay then. |
| `COOKIE_SECURE` | `false` | Set to `true` once the app is served over HTTPS. Adds the `Secure` flag and the `__Host-` prefix to the session cookie. |

## Sign-in and sessions

| Variable | Default | Description |
| --- | --- | --- |
| `SESSION_TTL` | `720h` | How long a session stays valid without use. Each use extends it. |
| `SETUP_TOKEN` | random, printed to the log | A fixed setup token of at least 16 characters, for automated installations. |
| `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` | empty | OpenID Connect sign-in for the owner. See [single sign-on](reverse-proxy.md#single-sign-on-optional). |
| `OIDC_ALLOWED_EMAILS`, `OIDC_ALLOWED_SUBJECTS` | empty | Who may sign in through OIDC, by verified e-mail address or subject. |
| `OIDC_GROUPS_CLAIM` | empty | ID token claim with a person's groups, e.g. `groups`. Used by [single sign-on for members](single-sign-on-members.md). |
| `TEAMS_BOT_APP_ID` | empty | Licensed `ee` builds: the organisation's Azure Bot app ID; switches on the Teams command box search at `POST /api/teams/messages`, which Microsoft must reach. See [Microsoft Teams](teams.md). |
| `SLACK_APP_TOKEN` | empty | Licensed `ee` builds: the app-level token (`xapp-…`) of the workspace's own Slack app; switches on the `/home` command over Socket Mode. See [Slack command](slack-command.md). |
| `OIDC_MEMBER_EMAILS`, `OIDC_MEMBER_GROUPS`, `OIDC_SYNC_GROUPS`, `OIDC_ADMIN_GROUPS` | empty | Licensed `ee` builds: who becomes a member or an admin through OIDC, and which provider groups are synced. See [single sign-on for members](single-sign-on-members.md). |
| `SCIM_TOKEN` | empty | Licensed `ee` builds: bearer token for SCIM provisioning from a directory. See [directory sync](directory-sync.md). |
| `LDAP_URL`, `LDAP_BIND_DN`, `LDAP_BIND_PASSWORD`, `LDAP_BASE_DN`, `LDAP_USER_ATTRIBUTE`, `LDAP_USER_CLASS`, `LDAP_MEMBER_GROUPS`, `LDAP_SYNC_GROUPS`, `LDAP_ADMIN_GROUPS`, `LDAP_CA_FILE` | empty | Licensed `ee` builds: members sign in with their LDAP or Active Directory password; `LDAP_ADMIN_GROUPS` makes some of them admins. See [directory sync](directory-sync.md#sign-in-with-ldap-or-active-directory). |
| `AUTH_PROXY_HEADER`, `AUTH_PROXY_USERS` | empty | Trust a user header set by a proxy in `TRUSTED_PROXIES` (forward auth). |
| `PASSWORD_LOGIN` | `true` | Set to `false` to turn off the owner's password sign-in once single sign-on works. |
| `SSO_REQUIRE_2FA` | `false` | Also ask for the local two-factor code after single sign-on. |
| `OIDC_BUTTON_LABEL` | empty | The provider's name on the sign-in button, such as `Microsoft` ("Sign in with Microsoft"), at most 40 characters. Empty keeps "Sign in with single sign-on". |
| `OIDC_BUTTON_ICON` | empty | An icon on that button: `microsoft` or `key`. |
| `OIDC_AUTO_REDIRECT` | `false` | With `PASSWORD_LOGIN=false`, the sign-in page goes straight to the provider. `/?local=1` still shows it. See [single sign-on](reverse-proxy.md#single-sign-on-optional). |
| `MICROSOFT_CLIENT_ID` | built in | Microsoft app registration for signing in with a personal Microsoft account by code, without any setup or HTTPS. Replace it with your own public client id; empty turns it off. See [sign in with Microsoft](sign-in-with-microsoft.md). |
| `MICROSOFT_ALLOWED_TENANTS` | empty | Directory (tenant) IDs, separated by commas, whose work or school accounts may sign in by code as well. Empty allows personal accounts only. See [sign in with Microsoft](sign-in-with-microsoft.md#by-code). |

## Actions that change other systems

All of these are off by default. Each one lists exactly what may be changed;
see the [host integrations](host-integrations.md) and the
[security model](security-model.md) for why.

| Variable | Default | Description |
| --- | --- | --- |
| `DOCKER_ACTIONS` | `false` | Enables restart and the log view in the Docker widget. Needs `COMPOSE_PROFILES=docker-actions`. |
| `DOCKER_ACTIONS_CONTAINERS` | empty | Exact names of the containers that may be restarted and whose logs may be shown. |
| `DOCKER_ACTIONS_PROXY_URL` | `http://docker-actions-proxy:2375` | The restart-only Docker socket proxy. |
| `SMARTHOME_ACTIONS_ENTITIES` | empty | Home Assistant entity IDs that may be switched or pressed, such as `switch.coffee_machine`. |
| `PROXMOX_ACTIONS_GUESTS` | empty | Proxmox VM and container IDs that may be started and shut down, such as `100,105`. |
| `DNS_PAUSE_PROVIDERS` | empty | DNS filters whose blocking may be paused for 5, 30 or 60 minutes: `adguard`, `pihole` or both. Refused together with `DEMO_MODE`. |

## Host data

| Variable | Default | Description |
| --- | --- | --- |
| `SYSTEM_DISK_PATHS` | `System=/host/rootfs` in Compose | Disks shown by the system widget, as comma-separated `Label=path` pairs. |
| `HOST_PROC` | `/proc`, `/host/proc` in Compose | The host's procfs, read by the system module. |
| `HOST_SYS` | `/sys`, `/host/sys` in Compose | The host's sysfs, for temperatures. |
| `DOCKER_PROXY_URL` | `http://docker-socket-proxy:2375` | The read-only Docker socket proxy. |
| `WIREGUARD_DIR` | `/wireguard` in Compose | Directory with the export of `scripts/wireguard-export.sh`. |
| `DISKHEALTH_DIR` | `/smart` in Compose | Directory with the export of `scripts/smart-export.sh`. |
| `DISKHEALTH_STALE_AFTER` | `2h` | S.M.A.R.T. exports older than this are shown as unknown. |
| `WOL_RELAY` | `host.docker.internal:9009` in Compose | The Wake-on-LAN relay. Empty sends the broadcast directly, for installations without Docker or in the host network. |
| `WOL_BROADCAST` | `255.255.255.255:9` | Where the magic packet is sent, by the relay or directly. |
| `DEVICE_DIR` | `/var/lib/waylight/device` | Raspberry Pi image only: where the chosen display mode is handed to the system service that switches the screen. |

## External services

Change these only to use a mirror, your own instance or a test server.

| Variable | Default | Description |
| --- | --- | --- |
| `TRANSIT_API_URL` | `https://api.transitous.org` | MOTIS API for departures. Point it to your own MOTIS instance if you run one. |
| `FUEL_API_URL` | `https://creativecommons.tankerkoenig.de/json` | Tankerkönig API for fuel prices. |
| `SPACE_LAUNCH_API_URL` | `https://ll.thespacedevs.com/2.2.0/launch/upcoming/` | Launch Library 2 for the space widget. |
| `SPACE_APOD_API_URL` | `https://api.nasa.gov/planetary/apod` | NASA's Astronomy Picture of the Day API. |

## Storage and backups

| Variable | Default | Description |
| --- | --- | --- |
| `DATABASE_PATH` | `/data/waylight.db` | The SQLite database file. |
| `BACKUP_DIR` | `backups` next to the database | Directory for the automatic backups. |
| `BACKUP_INTERVAL` | `24h` | How often a backup is made; at least `1h`. |
| `BACKUP_RETENTION` | `7` | How many backups are kept. |

## Operations

| Variable | Default | Description |
| --- | --- | --- |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. |
| `LOG_FORMAT` | `json` | `json` or `text`. |
| `SHUTDOWN_TIMEOUT` | `15s` | How long requests in flight may finish when the server stops. |
| `METRICS_ENABLED` | `false` | Expose `GET /metrics` for Prometheus. See [metrics](metrics.md). |
| `METRICS_TOKEN` | none | Bearer token for `/metrics`, at least 32 characters. |
| `INSTALL_METHOD` | set by the image, the package and the Pi image | `docker`, `package` or `appliance`. Chooses which update instructions the settings show. |
| `WAYLIGHT_VERSION` | `latest` | The image tag Docker Compose runs. A version such as `1.2.3` pins it. Docker Compose only. |
| `COMPOSE_PROFILES` | empty | Optional Compose services: `wol` for the Wake-on-LAN relay, `docker-actions` for the restart proxy. Docker Compose only. |

## Look

| Variable | Default | Description |
| --- | --- | --- |
| `DEFAULT_THEME` | `amber` | Id of a built-in theme, such as `amber-dark` or `nord`, for the sign-in page and every account that has not chosen its own appearance. `none` keeps the plain built-in look; an unknown id logs a warning and keeps it too. See [appearance](appearance.md#default-theme-for-everyone). |

## Sample data and demo

| Variable | Default | Description |
| --- | --- | --- |
| `SAMPLE_WEATHER_LOCATION` | `Berlin,52.52,13.41` | `Name,latitude,longitude` of the example weather widget. |
| `SAMPLE_FEEDS` | Go Blog, Hacker News | Example feeds, comma separated. |
| `DEMO_MODE` | `false` | A public demo without sign-in, with sample data and canned answers instead of real integrations. See [demo mode](demo.md). |
| `DEMO_RESET_INTERVAL` | `3h` | How often a demo restores its sample data. |
