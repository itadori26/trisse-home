# Drop zone

**What it does:** Share text, links and small files between your devices (#101). Paste or drop something on one device and pick it up on another. Items are listed newest first with copy and download buttons and expire automatically after 1 hour, 24 hours (default) or 7 days.

## Limits

Text up to 100 KB, files up to 25 MB, together at most 200 MB and 100 items. Expired items are never listed or served and are deleted within a quarter of an hour.

## Security

Everything stays on your server. Files are always offered as a **download** (`Content-Disposition: attachment`, `nosniff`, sandbox CSP). Only passive types (PNG, JPEG, GIF, WebP, PDF, plain text) keep their content type; everything else, HTML and SVG included, is sent as `application/octet-stream`. File names are reduced to a safe base name. Items are not part of data exports (they are short-lived); backups contain them.

## Outbound requests

None.

## API

`GET /api/drop`, `POST /api/drop/text`, `POST /api/drop/files` (multipart), `DELETE /api/drop/{id}`, `GET /api/drop/{id}/file` – see `api/modules/drop.yaml`.
