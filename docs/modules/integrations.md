# Integrations

**What it does:** an admin saves the address and the credential of a
service once, as a named connection such as "Home" or "Office", and widgets
of that service choose it. A service can have several connections, for
example two Proxmox clusters.

## Status

This is the store and its API. Widgets and the settings pages of the
integrations do not use it yet; their own settings stay as they are until
they move to connections (#610).

## Reading through a connection

Modules that read through connections take the connection a widget chose
(`?connection=<id>` on their read endpoints). While that connection does
not exist, belongs to another service or lacks a required field (for
example after a `SECRET_KEY` change), the module uses its own settings as
before. Saving in a module's settings page still writes the module
settings only.

The background jobs of Nextcloud, Proxmox VE and Speedtest Tracker refresh
the module settings only. A connection is read when a widget asks, at most
as often as the job runs (5 minutes, 30 seconds and 10 minutes).

| Module | Read endpoints |
| --- | --- |
| Immich | `GET /api/immich/overview`, `GET /api/immich/thumbnails/{id}` |
| Home Assistant | `GET /api/smarthome/entities`, `GET /api/smarthome/states`, `GET /api/smarthome/compact`, `POST /api/smarthome/entities/{id}/action` (acts on the same server) |
| Nextcloud | `GET /api/nextcloud/status` |
| Proxmox VE | `GET /api/proxmox/status`, `GET /api/proxmox/compact`, `POST /api/proxmox/guests/{id}/{action}` (acts on the same cluster) |
| Speedtest Tracker | `GET /api/speedtest/results`, `GET /api/speedtest/compact` (the thresholds stay in the module settings) |
| Grafana | `GET /api/grafana/alerts`, `GET /api/grafana/dashboards`, `GET /api/grafana/dashboards/{uid}/panels`, `GET /api/grafana/render/{uid}/{panel}`, `GET /api/grafana/data/{uid}/{panel}` |

## Copy of the module settings

At start and then every hour, the settings of the integration modules are
copied into connections, each once:

- Immich, Grafana, Nextcloud, Proxmox VE, Home Assistant, Speedtest
  Tracker, the FRITZ!Box and the media server become one connection named
  after the product; AdGuard Home and Pi-hole one each; every Sonarr,
  Radarr and Prowlarr instance one with its own name; the Tankerkönig key
  one with the fixed address. Settings that are not credentials (place and
  radius of fuel prices, Speedtest thresholds) stay in the module.
- A name another connection of the service already has gets a number
  ("Immich 2").
- Each copy is read back and decrypted in the same transaction and compared
  with the module settings. Only then is it kept and recorded; otherwise
  nothing is kept and the next run tries again. The log names the source
  and the reason, never a credential.
- A recorded source is not copied again, also when its connection is
  removed later. Settings saved in a module after the first run are copied
  by the next run.
- The module settings are only read. Widgets keep using them until they
  read connections.
- While [config as code](../config-as-code.md) manages the `integrations`
  section, nothing is copied: the file stays the only source of the list.

## Services

Each integration module declares the services it reaches and the
credential fields they need. `GET /api/integrations/services` lists them:

| Service | Product | Fields (secret in bold, optional in italics) |
| --- | --- | --- |
| `adguard` | AdGuard Home | `username`, **`password`** |
| `fritzbox` | FRITZ!Box | *`username`*, ***`password`*** |
| `grafana` | Grafana | **`token`** |
| `homeassistant` | Home Assistant | **`token`** |
| `immich` | Immich | **`apiKey`** |
| `jellyfin`, `plex` | Jellyfin, Plex | **`token`** |
| `nextcloud` | Nextcloud | `username`, **`appPassword`** |
| `pihole` | Pi-hole | *`username`*, **`password`** |
| `proxmox` | Proxmox VE | `tokenId`, **`secret`**, *`fingerprint`* |
| `sonarr`, `radarr`, `prowlarr` | Sonarr, Radarr, Prowlarr | **`apiKey`** |
| `speedtest` | Speedtest Tracker | **`token`** |
| `tankerkoenig` | Tankerkönig | **`apiKey`**; the address is fixed |

A new connection must name one of these services and fill its required
fields; fields the service does not know are refused. Connections of a
service whose module is switched off stay in the list and editable.

## Rules

- Only admins see and change connections; they belong to the instance.
- A name is unique per service, compared without case. An instance keeps
  at most 100 connections.
- The credential is a set of named fields (for example `apiKey`, or
  `username` and `password`), at most four. It is stored encrypted with
  `SECRET_KEY`. Secret fields are never returned: the API says which fields
  are set and shows the values of the others (`values`), such as a user name.
- Moving a connection to another address needs the credential again, so a
  stored one is never sent to a new host.
- Removing a connection that widgets name in their settings
  (`connectionId`) is refused until it is confirmed; the list shows which
  widgets these are.
- If `SECRET_KEY` changes, the credentials can no longer be read. The list
  still opens and shows no fields; enter the credentials again.

## Checking a connection

`POST /api/integrations/connections/{id}/check` sends the request the
service's settings page makes before saving, with the stored address and
credential. It answers 204 when the service accepts it and otherwise the
same message the settings page shows.

## Config as code

The `integrations` section of [waylight.yaml](../config-as-code.md) lists
service, name and address. It never contains a credential. While the file
manages the section, connections cannot be added, renamed or removed in the
app, but credentials can still be entered.

## API

`GET /api/integrations/connections` (optionally `?service=immich`)
lists the connections with `fields`, `values` and `usedBy`;
`GET /api/integrations/services` lists the services; `POST /api/integrations/connections`,
`PUT /api/integrations/connections/{id}`,
`PUT /api/integrations/connections/{id}/credential` and
`DELETE /api/integrations/connections/{id}?confirm=true` change them. See
`api/modules/integrations.yaml`.

## Outbound requests

Only the check: one request to the connection's address, the same one the
service's settings page makes.
