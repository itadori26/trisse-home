# Space

**What it does:** The next rocket launches (rocket, mission, local time, status) from [Launch Library 2](https://thespacedevs.com/llapi) and NASA's [Astronomy Picture of the Day](https://apod.nasa.gov) with title and explanation (#64).

## Configuration

Settings → Space: switch launches and/or the picture of the day on (both are off by default). Optionally enter your own free [NASA API key](https://api.nasa.gov); without one the public `DEMO_KEY` is used, which has a low rate limit per IP address. The key is stored encrypted and never returned.

**Picture as background:** the server serves today's picture at the stable address `/api/space/apod/image` (the image is loaded by the server, so your browser does not contact NASA). Use the full address, e.g. `http://homeserver:8090/api/space/apod/image`, as tab background (Edit → 🖼) or as the global background (Settings → Appearance); it changes every day. On days with a video there is no image.

## Outbound requests

Only after switching a source on, and only while a widget shows it (or the background is in use): Launch Library 2 at most once an hour (its free limit is 15 requests an hour), NASA APOD at most every three hours, and the day's image once. Images are only fetched from NASA hosts over https. `SPACE_LAUNCH_API_URL` and `SPACE_APOD_API_URL` change the APIs (only for a mirror or tests).

## API

`GET /api/space/launches?limit=5`, `GET /api/space/apod`, `GET /api/space/apod/image`, `GET/PUT /api/space/settings` – see `api/modules/space.yaml`.
