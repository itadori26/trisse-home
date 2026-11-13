# *arr (Sonarr, Radarr, Prowlarr)

**What it does:** Sonarr, Radarr and Prowlarr in one widget (several instances): upcoming episodes and movies for two weeks, download queue with progress and stuck downloads, health warnings and the number of missing items.

## Configuration

Settings → *arr: add each instance with its type, URL and API key. The keys are stored encrypted and never returned by the API.

## Outbound requests

The configured instances only, while an *arr widget is open, at most once a minute (calendar, queue, health and missing-item counts per instance).

## API

`GET/POST /api/arr/instances`, `DELETE /api/arr/instances/{id}`, `GET /api/arr/overview` – see `api/modules/arr.yaml`.
