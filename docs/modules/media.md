# Media server (Jellyfin, Plex)

**What it does:** Jellyfin or Plex: what is playing right now (user, device, progress) and recently added movies, episodes and albums with posters. Posters are proxied by trisse-home, so the token stays on the server.

## Configuration

Settings → Media server: provider, URL and token (stored encrypted, never returned by the API).

## Outbound requests

The configured server only, while a media widget is open, at most every 15 seconds (sessions, recently added items and the posters of those items).

## API

`GET/PUT /api/media/settings`, `GET /api/media/overview`, `GET /api/media/posters/{id}` – see `api/modules/media.yaml`.
