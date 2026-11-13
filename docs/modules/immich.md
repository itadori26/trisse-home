# Immich

**What it does:** Immich photo library on the dashboard: number of photos and videos, storage, and today's "on this day" memories. Thumbnails are proxied by trisse-home, so the API key never reaches the browser.

## Configuration

Settings → Immich: the server URL (e.g. `http://192.168.1.2:2283`) and an API key. Create the key under Account settings → API keys with the permissions `asset.statistics`, `server.storage`, `memory.read` and `asset.view`. The key is stored encrypted. Storage and memories are left out when the key lacks `server.storage` or `memory.read`.

## Outbound requests

Only the configured Immich server, while an Immich widget is open, at most every 5 minutes (`/api/assets/statistics`, `/api/server/version`, `/api/server/storage`, `/api/memories`, thumbnails of today's memories). The API key is sent as `x-api-key`.

## API

`GET/PUT /api/immich/settings`, `GET /api/immich/overview`, `GET /api/immich/thumbnails/{id}` – see `api/openapi.yaml`.
